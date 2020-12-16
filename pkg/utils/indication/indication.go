// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package indication

import (
	"github.com/onosproject/onos-e2t/api/e2ap/v1beta1"
	e2ap_commondatatypes "github.com/onosproject/onos-e2t/api/e2ap/v1beta1/e2ap-commondatatypes"
	"github.com/onosproject/onos-e2t/api/e2ap/v1beta1/e2apies"
	"github.com/onosproject/onos-e2t/api/e2ap/v1beta1/e2appducontents"
)

// Indication indication data struct
type Indication struct {
	reqID             int32
	ricInstanceID     int32
	ranFuncID         int32
	indicationHeader  []byte
	indicationMessage []byte
	ricCallProcessID  []byte
}

// NewIndication creates a new indication
func NewIndication(options ...func(*Indication)) (*Indication, error) {
	indication := &Indication{}

	for _, option := range options {
		option(indication)
	}

	return indication, nil

}

// WithRequestID sets request ID
func WithRequestID(reqID int32) func(*Indication) {
	return func(indication *Indication) {
		indication.reqID = reqID
	}
}

// WithRanFuncID sets ran function ID
func WithRanFuncID(ranFuncID int32) func(*Indication) {
	return func(indication *Indication) {
		indication.ranFuncID = ranFuncID
	}
}

// WithRicInstanceID sets ric instance ID
func WithRicInstanceID(ricInstanceID int32) func(*Indication) {
	return func(indication *Indication) {
		indication.ricInstanceID = ricInstanceID
	}
}

// WithIndicationHeader sets indication header
func WithIndicationHeader(indicationHeader []byte) func(*Indication) {
	return func(indication *Indication) {
		indication.indicationHeader = indicationHeader
	}
}

// WithIndicationMessage sets indication message
func WithIndicationMessage(indicationMessage []byte) func(*Indication) {
	return func(indication *Indication) {
		indication.indicationMessage = indicationMessage
	}
}

// CreateIndication creates indication message
func CreateIndication(indication *Indication) (e2Indication *e2appducontents.Ricindication) {
	ricIndication := &e2appducontents.Ricindication{
		ProtocolIes: &e2appducontents.RicindicationIes{
			E2ApProtocolIes29: &e2appducontents.RicindicationIes_RicindicationIes29{
				Id:          int32(v1beta1.ProtocolIeIDRicrequestID),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value: &e2apies.RicrequestId{
					RicRequestorId: indication.reqID,
					RicInstanceId:  indication.ricInstanceID,
				},
				Presence: int32(e2ap_commondatatypes.Presence_PRESENCE_MANDATORY),
			},
			E2ApProtocolIes5: &e2appducontents.RicindicationIes_RicindicationIes5{
				Id:          int32(v1beta1.ProtocolIeIDRanfunctionID),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value: &e2apies.RanfunctionId{
					Value: indication.ranFuncID,
				},
				Presence: int32(e2ap_commondatatypes.Presence_PRESENCE_MANDATORY),
			},
			E2ApProtocolIes15: &e2appducontents.RicindicationIes_RicindicationIes15{
				Id:          int32(v1beta1.ProtocolIeIDRicactionID),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value: &e2apies.RicactionId{
					Value: 2,
				},
				Presence: int32(e2ap_commondatatypes.Presence_PRESENCE_MANDATORY),
			},
			E2ApProtocolIes27: &e2appducontents.RicindicationIes_RicindicationIes27{
				Id:          int32(v1beta1.ProtocolIeIDRicindicationSn),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value: &e2apies.RicindicationSn{
					Value: 3,
				},
				Presence: int32(e2ap_commondatatypes.Presence_PRESENCE_OPTIONAL),
			},
			E2ApProtocolIes28: &e2appducontents.RicindicationIes_RicindicationIes28{
				Id:          int32(v1beta1.ProtocolIeIDRicindicationType),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value:       e2apies.RicindicationType_RICINDICATION_TYPE_REPORT,
				Presence:    int32(e2ap_commondatatypes.Presence_PRESENCE_MANDATORY),
			},
			E2ApProtocolIes25: &e2appducontents.RicindicationIes_RicindicationIes25{
				Id:          int32(v1beta1.ProtocolIeIDRicindicationHeader),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value: &e2ap_commondatatypes.RicindicationHeader{
					Value: indication.indicationHeader,
				},
				Presence: int32(e2ap_commondatatypes.Presence_PRESENCE_MANDATORY),
			},
			E2ApProtocolIes26: &e2appducontents.RicindicationIes_RicindicationIes26{
				Id:          int32(v1beta1.ProtocolIeIDRicindicationMessage),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value: &e2ap_commondatatypes.RicindicationMessage{
					Value: indication.indicationMessage,
				},
				Presence: int32(e2ap_commondatatypes.Presence_PRESENCE_MANDATORY),
			},
			E2ApProtocolIes20: &e2appducontents.RicindicationIes_RicindicationIes20{
				Id:          int32(v1beta1.ProtocolIeIDRiccallProcessID),
				Criticality: int32(e2ap_commondatatypes.Criticality_CRITICALITY_REJECT),
				Value: &e2ap_commondatatypes.RiccallProcessId{
					Value: indication.ricCallProcessID,
				},
				Presence: int32(e2ap_commondatatypes.Presence_PRESENCE_OPTIONAL),
			},
		},
	}

	return ricIndication

}