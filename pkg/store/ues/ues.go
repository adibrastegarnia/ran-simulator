// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package ues

import (
	"github.com/onosproject/onos-lib-go/pkg/errors"
	liblog "github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/ran-simulator/api/types"
	"github.com/onosproject/ran-simulator/pkg/model"
	"github.com/onosproject/ran-simulator/pkg/store/cells"
	"math/rand"
	"sync"
)

const (
	minIMSI = 1000000
	maxIMSI = 9999999
)

var log = liblog.GetLogger("store", "ues")

// UERegistry tracks inventory of user-equipment for the simulation
type UERegistry interface {
	// SetUECount updates the UE count and creates or deletes new UEs as needed
	SetUECount(count uint)

	// GetUECount returns the number of active UEs
	GetUECount() uint

	// CreateUEs creates the specified number of UEs
	CreateUEs(count uint)

	// GetUE retrieves the UE with the specified IMSI
	GetUE(imsi types.IMSI) (*model.UE, error)

	// DestroyUE destroy the specified UE
	DestroyUE(imsi types.IMSI) (*model.UE, error)

	// MoveUE update the cell affiliation of the specified UE
	MoveUE(imsi types.IMSI, ecgi types.ECGI, strength float64) error

	// ListAllUEs returns an array of all UEs
	ListAllUEs() []*model.UE

	// ListUEs returns an array of all UEs associated with the specified cell
	ListUEs(ecgi types.ECGI) []*model.UE

	// WatchUEs watches the UE inventory events using the supplied channel
	WatchUEs(ch chan<- UEEvent, options ...WatchOptions)
}

// UEEvent represents a change in the node inventory
type UEEvent struct {
	UE   *model.UE
	Type uint8
}

// WatchOptions allows tailoring the WatchUEs behaviour
type WatchOptions struct {
	Replay  bool
	Monitor bool
}

type ueWatcher struct {
	ch chan<- UEEvent
}

func (r *ueRegistry) notify(ue *model.UE, eventType uint8) {
	event := UEEvent{
		UE:   ue,
		Type: eventType,
	}
	for _, watcher := range r.watchers {
		watcher.ch <- event
	}
}

type ueRegistry struct {
	lock      sync.RWMutex
	ues       map[types.IMSI]*model.UE
	watchers  []ueWatcher
	cellStore cells.CellRegistry
}

// NewUERegistry creates a new user-equipment registry primed with the specified number of UEs to start.
// UEs will be semi-randomly distributed between the specified cells
func NewUERegistry(count uint, cellStore cells.CellRegistry) UERegistry {
	log.Infof("Creating registry from model with %d UEs", count)
	reg := &ueRegistry{
		lock:      sync.RWMutex{},
		ues:       make(map[types.IMSI]*model.UE),
		cellStore: cellStore,
	}
	reg.CreateUEs(count)
	log.Infof("Created registry primed with %d UEs", len(reg.ues))
	return reg
}

func (r *ueRegistry) SetUECount(count uint) {
	delta := len(r.ues) - int(count)
	if delta < 0 {
		r.CreateUEs(uint(-delta))
	} else if delta > 0 {
		r.removeSomeUEs(delta)
	}
}

func (r *ueRegistry) GetUECount() uint {
	return uint(len(r.ues))
}

func (r *ueRegistry) removeSomeUEs(count int) {
	c := count
	for imsi := range r.ues {
		if c == 0 {
			break
		}
		_, _ = r.DestroyUE(imsi)
		c = c - 1
	}
}

func (r *ueRegistry) CreateUEs(count uint) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for i := uint(0); i < count; i++ {
		imsi := types.IMSI(rand.Int63n(maxIMSI-minIMSI) + minIMSI)
		if _, ok := r.ues[imsi]; ok {
			// FIXME: more robust check for duplicates
			imsi = types.IMSI(rand.Int63n(maxIMSI-minIMSI) + minIMSI)
		}

		ecgi := r.cellStore.GetRandomCell().ECGI
		ue := &model.UE{
			IMSI:     imsi,
			Type:     "phone",
			Location: model.Coordinate{Lat: 0, Lng: 0},
			Rotation: 0,
			Cell: &model.UECell{
				ID:       types.GEnbID(ecgi), // placeholder
				ECGI:     ecgi,
				Strength: rand.Float64() * 100,
			},
			CRNTI:      types.CRNTI(90125 + i),
			Cells:      nil,
			IsAdmitted: false,
		}
		r.ues[ue.IMSI] = ue
	}
}

func (r *ueRegistry) GetUE(imsi types.IMSI) (*model.UE, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if node, ok := r.ues[imsi]; ok {
		return node, nil
	}

	return nil, errors.New(errors.NotFound, "UE not found")
}

func (r *ueRegistry) DestroyUE(imsi types.IMSI) (*model.UE, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if ue, ok := r.ues[imsi]; ok {
		delete(r.ues, imsi)
		r.notify(ue, DELETED)
		return ue, nil
	}
	return nil, errors.New(errors.NotFound, "UE not found")
}

func (r *ueRegistry) ListAllUEs() []*model.UE {
	r.lock.RLock()
	defer r.lock.RUnlock()
	list := make([]*model.UE, 0, len(r.ues))
	for _, ue := range r.ues {
		list = append(list, ue)
	}
	return list
}

func (r *ueRegistry) MoveUE(imsi types.IMSI, ecgi types.ECGI, strength float64) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if ue, ok := r.ues[imsi]; ok {
		ue.Cell.ECGI = ecgi
		ue.Cell.Strength = strength
		r.notify(ue, UPDATED)
		return nil
	}
	return errors.New(errors.NotFound, "UE not found")
}

func (r *ueRegistry) ListUEs(ecgi types.ECGI) []*model.UE {
	r.lock.RLock()
	defer r.lock.RUnlock()
	list := make([]*model.UE, 0, len(r.ues))
	for _, ue := range r.ues {
		if ue.Cell.ECGI == ecgi {
			list = append(list, ue)
		}
	}
	return list
}

const (
	// NONE indicates no change event
	NONE uint8 = 0

	// ADDED indicates new node was added
	ADDED uint8 = 1

	// UPDATED indicates an existing node was updated
	UPDATED uint8 = 2

	// DELETED indicates a node was deleted
	DELETED uint8 = 3
)

func (r *ueRegistry) WatchUEs(ch chan<- UEEvent, options ...WatchOptions) {
	log.Infof("WatchUEs: %v (#%d)\n", options, len(r.ues))
	monitor := len(options) == 0 || options[0].Monitor
	replay := len(options) > 0 && options[0].Replay
	go func() {
		watcher := ueWatcher{ch: ch}
		if monitor {
			r.lock.RLock()
			r.watchers = append(r.watchers, watcher)
			r.lock.RUnlock()
		}

		if replay {
			r.lock.RLock()
			defer r.lock.RUnlock()
			for _, ue := range r.ues {
				ch <- UEEvent{UE: ue, Type: NONE}
			}
			if !monitor {
				close(ch)
			}
		}
	}()

}