// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package kpm

import (
	"context"
	"fmt"
	"strconv"
	"time"

	e2smtypes "github.com/onosproject/onos-api/go/onos/e2t/e2sm"

	ransimtypes "github.com/onosproject/onos-api/go/onos/ransim/types"

	"github.com/onosproject/ran-simulator/pkg/store/nodes"
	"github.com/onosproject/ran-simulator/pkg/store/ues"

	"github.com/onosproject/ran-simulator/pkg/store/subscriptions"

	kpmutils "github.com/onosproject/ran-simulator/pkg/utils/e2sm/kpm/indication"

	"github.com/onosproject/ran-simulator/pkg/model"

	"github.com/onosproject/ran-simulator/pkg/modelplugins"

	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_kpm/pdubuilder"
	indicationutils "github.com/onosproject/ran-simulator/pkg/utils/e2ap/indication"
	subutils "github.com/onosproject/ran-simulator/pkg/utils/e2ap/subscription"
	subdeleteutils "github.com/onosproject/ran-simulator/pkg/utils/e2ap/subscriptiondelete"

	"github.com/onosproject/onos-lib-go/pkg/logging"

	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/ran-simulator/pkg/servicemodel/registry"

	e2apies "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-ies"
	e2appducontents "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-pdu-contents"
	e2aptypes "github.com/onosproject/onos-e2t/pkg/southbound/e2ap/types"
	"github.com/onosproject/ran-simulator/pkg/servicemodel"
	"google.golang.org/protobuf/proto"
)

var _ servicemodel.Client = &Client{}

var log = logging.GetLogger("sm", "kpm")

const (
	modelName = "ORAN-E2SM-KPM"
	version   = "v1"
	modelOID  = "1.3.6.1.4.1.53148.1.1.2.2"
)

// Client kpm service model client
type Client struct {
	ServiceModel *registry.ServiceModel
}

// NewServiceModel creates a new service model
func NewServiceModel(node model.Node, model *model.Model, modelPluginRegistry modelplugins.ModelRegistry,
	subStore *subscriptions.Subscriptions, nodeStore nodes.Store, ueStore ues.Store) (registry.ServiceModel, error) {
	modelName := e2smtypes.ShortName(modelName)
	kpmSm := registry.ServiceModel{
		RanFunctionID:       registry.Kpm,
		ModelName:           modelName,
		Revision:            1,
		OID:                 modelOID,
		Version:             version,
		ModelPluginRegistry: modelPluginRegistry,
		Node:                node,
		Model:               model,
		Subscriptions:       subStore,
		Nodes:               nodeStore,
		UEs:                 ueStore,
	}
	kpmClient := &Client{
		ServiceModel: &kpmSm,
	}

	kpmSm.Client = kpmClient

	var ranFunctionShortName = string(modelName)
	var ranFunctionE2SmOid = modelOID
	var ranFunctionDescription = "KPM Monitor"
	var ranFunctionInstance int32 = 1
	var ricEventStyleType int32 = 1
	var ricEventStyleName = "Periodic report"
	var ricEventFormatType int32 = 5
	var ricReportStyleType int32 = 1
	var ricReportStyleName = "O-CU-CP Measurement Container for the 5GC connected deployment"
	var ricIndicationHeaderFormatType int32 = 1
	var ricIndicationMessageFormatType int32 = 1
	ranFuncDescPdu, err := pdubuilder.CreateE2SmKpmRanfunctionDescriptionMsg(ranFunctionShortName, ranFunctionE2SmOid, ranFunctionDescription,
		ranFunctionInstance, ricEventStyleType, ricEventStyleName, ricEventFormatType, ricReportStyleType, ricReportStyleName,
		ricIndicationHeaderFormatType, ricIndicationMessageFormatType)
	if err != nil {
		log.Error(err)
		return registry.ServiceModel{}, err
	}

	protoBytes, err := proto.Marshal(ranFuncDescPdu)
	if err != nil {
		log.Error(err)
		return registry.ServiceModel{}, err
	}

	kpmModelPlugin, err := modelPluginRegistry.GetPlugin(modelOID)
	if kpmModelPlugin == nil {
		return registry.ServiceModel{}, errors.New(errors.Invalid, "model plugin is nil: %v", err)
	}
	ranFuncDescBytes, err := kpmModelPlugin.RanFuncDescriptionProtoToASN1(protoBytes)
	if err != nil {
		log.Error(err)
		return registry.ServiceModel{}, err
	}

	kpmSm.Description = ranFuncDescBytes
	return kpmSm, nil
}

func (sm *Client) reportIndication(ctx context.Context, interval int32, subscription *subutils.Subscription) error {
	subID := subscriptions.NewID(subscription.GetRicInstanceID(), subscription.GetReqID(), subscription.GetRanFuncID())
	gNbID, err := strconv.ParseUint(fmt.Sprintf("%d", sm.ServiceModel.Node.GnbID), 10, 64)
	if err != nil {
		log.Error(err)
		return err
	}
	// Creates an indication header
	plmnID := ransimtypes.NewUint24(uint32(sm.ServiceModel.Model.PlmnID))

	header := kpmutils.NewIndicationHeader(
		kpmutils.WithPlmnID(plmnID.Value()),
		kpmutils.WithGnbID(gNbID),
		kpmutils.WithSst("1"),
		kpmutils.WithSd("SD1"),
		kpmutils.WithPlmnIDnrcgi(plmnID.Value()))

	kpmModelPlugin, _ := sm.ServiceModel.ModelPluginRegistry.GetPlugin(e2smtypes.OID(sm.ServiceModel.OID))
	indicationHeaderAsn1Bytes, err := header.ToAsn1Bytes(kpmModelPlugin)
	if err != nil {
		log.Error(err)
		return err
	}

	// Creating an indication message
	indicationMessage := kpmutils.NewIndicationMessage(
		kpmutils.WithNumberOfActiveUes(int32(sm.ServiceModel.UEs.Len(ctx))))

	indicationMessageBytes, err := indicationMessage.ToAsn1Bytes(kpmModelPlugin)
	if err != nil {
		log.Error(err)
		return err
	}

	intervalDuration := time.Duration(interval)
	sub, err := sm.ServiceModel.Subscriptions.Get(subID)
	if err != nil {
		log.Error(err)
		return err
	}
	sub.Ticker = time.NewTicker(intervalDuration * time.Millisecond)
	for {
		select {
		case <-sub.Ticker.C:
			log.Debug("Sending Indication Report for subscription:", sub.ID)
			indication := indicationutils.NewIndication(
				indicationutils.WithRicInstanceID(subscription.GetRicInstanceID()),
				indicationutils.WithRanFuncID(subscription.GetRanFuncID()),
				indicationutils.WithRequestID(subscription.GetReqID()),
				indicationutils.WithIndicationHeader(indicationHeaderAsn1Bytes),
				indicationutils.WithIndicationMessage(indicationMessageBytes))

			ricIndication, err := indication.Build()
			if err != nil {
				log.Error("creating indication message is failed", err)
				return err
			}

			err = sub.E2Channel.RICIndication(ctx, ricIndication)
			if err != nil {
				log.Error("Sending indication report is failed:", err)
				return err
			}

		case <-sub.E2Channel.Context().Done():
			log.Debug("E2 channel context is done")
			sub.Ticker.Stop()
			return nil

		}
	}
}

// E2ConnectionUpdate implements connection update handler
func (sm *Client) E2ConnectionUpdate(ctx context.Context, request *e2appducontents.E2ConnectionUpdate) (response *e2appducontents.E2ConnectionUpdateAcknowledge, failure *e2appducontents.E2ConnectionUpdateFailure, err error) {
	return nil, nil, errors.NewNotSupported("E2 connection update is not supported")
}

// RICControl implements control handler for kpm service model
func (sm *Client) RICControl(ctx context.Context, request *e2appducontents.RiccontrolRequest) (response *e2appducontents.RiccontrolAcknowledge, failure *e2appducontents.RiccontrolFailure, err error) {
	return nil, nil, errors.New(errors.NotSupported, "Control operation is not supported")
}

// RICSubscription implements subscription handler for kpm service model
func (sm *Client) RICSubscription(ctx context.Context, request *e2appducontents.RicsubscriptionRequest) (response *e2appducontents.RicsubscriptionResponse, failure *e2appducontents.RicsubscriptionFailure, err error) {
	log.Infof("RIC Subscription request received for e2 node %d and service model %s:", sm.ServiceModel.Node.GnbID, sm.ServiceModel.ModelName)
	var ricActionsAccepted []*e2aptypes.RicActionID
	ricActionsNotAdmitted := make(map[e2aptypes.RicActionID]*e2apies.Cause)
	actionList := subutils.GetRicActionToBeSetupList(request)
	reqID, err := subutils.GetRequesterID(request)
	if err != nil {
		return nil, nil, err
	}
	ranFuncID, err := subutils.GetRanFunctionID(request)
	if err != nil {
		return nil, nil, err
	}
	ricInstanceID, err := subutils.GetRicInstanceID(request)
	if err != nil {
		return nil, nil, err
	}

	for _, action := range actionList {
		actionID := e2aptypes.RicActionID(action.GetValue().GetRatbsi().GetRicActionId().GetValue())
		actionType := action.GetValue().GetRatbsi().GetRicActionType()
		// kpm service model supports report action and should be added to the
		// list of accepted actions
		if actionType == e2apies.RicactionType_RICACTION_TYPE_REPORT {
			ricActionsAccepted = append(ricActionsAccepted, &actionID)
		}
		// kpm service model does not support INSERT and POLICY actions and
		// should be added into the list of not admitted actions
		if actionType == e2apies.RicactionType_RICACTION_TYPE_INSERT ||
			actionType == e2apies.RicactionType_RICACTION_TYPE_POLICY {
			cause := &e2apies.Cause{
				Cause: &e2apies.Cause_RicRequest{
					RicRequest: e2apies.CauseRicrequest_CAUSE_RICREQUEST_ACTION_NOT_SUPPORTED,
				},
			}
			ricActionsNotAdmitted[actionID] = cause
		}
	}

	// At least one required action must be accepted otherwise sends a subscription failure response
	if len(ricActionsAccepted) == 0 {
		log.Warn("no action is accepted")
		cause := &e2apies.Cause{
			Cause: &e2apies.Cause_RicRequest{
				RicRequest: e2apies.CauseRicrequest_CAUSE_RICREQUEST_ACTION_NOT_SUPPORTED,
			},
		}
		subscription := subutils.NewSubscription(
			subutils.WithRequestID(*reqID),
			subutils.WithRanFuncID(*ranFuncID),
			subutils.WithRicInstanceID(*ricInstanceID),
			subutils.WithCause(cause))
		subscriptionFailure, err := subscription.BuildSubscriptionFailure()
		if err != nil {
			return nil, nil, err
		}
		return nil, subscriptionFailure, nil
	}

	reportInterval, err := sm.getReportPeriod(request)
	if err != nil {
		log.Warn(err)
		cause := &e2apies.Cause{
			Cause: &e2apies.Cause_RicRequest{
				RicRequest: e2apies.CauseRicrequest_CAUSE_RICREQUEST_UNSPECIFIED,
			},
		}
		subscription := subutils.NewSubscription(
			subutils.WithRequestID(*reqID),
			subutils.WithRanFuncID(*ranFuncID),
			subutils.WithRicInstanceID(*ricInstanceID),
			subutils.WithCause(cause))
		subscriptionFailure, err := subscription.BuildSubscriptionFailure()
		if err != nil {
			log.Warn(err)
			return nil, subscriptionFailure, nil
		}
		return nil, subscriptionFailure, nil
	}

	subscription := subutils.NewSubscription(
		subutils.WithRequestID(*reqID),
		subutils.WithRanFuncID(*ranFuncID),
		subutils.WithRicInstanceID(*ricInstanceID),
		subutils.WithActionsAccepted(ricActionsAccepted),
		subutils.WithActionsNotAdmitted(ricActionsNotAdmitted))
	subscriptionResponse, err := subscription.BuildSubscriptionResponse()
	if err != nil {
		return nil, nil, err
	}
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := sm.reportIndication(ctx, reportInterval, subscription)
		if err != nil {
			return
		}
	}()
	return subscriptionResponse, nil, nil

}

// RICSubscriptionDelete implements subscription delete handler for kpm service model
func (sm *Client) RICSubscriptionDelete(ctx context.Context, request *e2appducontents.RicsubscriptionDeleteRequest) (response *e2appducontents.RicsubscriptionDeleteResponse, failure *e2appducontents.RicsubscriptionDeleteFailure, err error) {
	log.Infof("RIC subscription delete request is received for e2 node %d and  service model %s:", sm.ServiceModel.Node.GnbID, sm.ServiceModel.ModelName)
	reqID, err := subdeleteutils.GetRequesterID(request)
	if err != nil {
		return nil, nil, err
	}
	ranFuncID, err := subdeleteutils.GetRanFunctionID(request)
	if err != nil {
		return nil, nil, err
	}
	ricInstanceID, err := subdeleteutils.GetRicInstanceID(request)
	if err != nil {
		return nil, nil, err
	}
	subID := subscriptions.NewID(*ricInstanceID, *reqID, *ranFuncID)
	sub, err := sm.ServiceModel.Subscriptions.Get(subID)
	if err != nil {
		return nil, nil, err
	}
	subscriptionDelete := subdeleteutils.NewSubscriptionDelete(
		subdeleteutils.WithRequestID(*reqID),
		subdeleteutils.WithRanFuncID(*ranFuncID),
		subdeleteutils.WithRicInstanceID(*ricInstanceID))
	subDeleteResponse, err := subscriptionDelete.BuildSubscriptionDeleteResponse()
	if err != nil {
		return nil, nil, err
	}
	// Stops the goroutine sending the indication messages
	sub.Ticker.Stop()
	return subDeleteResponse, nil, nil
}
