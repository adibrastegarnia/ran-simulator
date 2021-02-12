// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package model

// PlmnID is a globally unique network identifier (Public Land Mobile Network)
type PlmnID uint32

// EnbID is an eNodeB Identifier
type EnbID uint32

// CellID is a node-local cell identifier
type CellID uint8

// ECI is a E-UTRAN Cell Identifier
type ECI uint32

// GEnbID is a Globally eNodeB identifier
type GEnbID uint64

// ECGI is E-UTRAN Cell Global Identifier
type ECGI uint64

// CRNTI is a cell-specific UE identifier
type CRNTI uint32

// MSIN is Mobile Subscriber Identification Number
type MSIN uint32

// IMSI is International Mobile Subscriber Identity
type IMSI uint64

// Coordinate represents a geographical location
type Coordinate struct {
	Lat float64 `yaml:"lat"`
	Lng float64 `yaml:"lng"`
}

// Sector represents a 2D arc emanating from a location
type Sector struct {
	Center  Coordinate `yaml:"center"`
	Azimuth int32      `yaml:"azimuth"`
	Arc     int32      `yaml:"arc"`
}

// Route represents a named series of points for tracking movement of user-equipment
type Route struct {
	Name   string
	Points []*Coordinate
	Color  string
}

const mask28 = 0xfffffff
const mask20 = 0xfffff00

// ToECI produces ECI from the specified components
func ToECI(enbID EnbID, cid CellID) ECI {
	return ECI(uint(enbID)<<8 | uint(cid))
}

// ToECGI produces ECGI from the specified components
func ToECGI(plmnID PlmnID, eci ECI) ECGI {
	return ECGI(uint(plmnID)<<28 | (uint(eci) & mask28))
}

// ToGEnbID produces GEnbID from the specified components
func ToGEnbID(plmnID PlmnID, enbID EnbID) GEnbID {
	return GEnbID(uint(plmnID)<<28 | (uint(enbID) << 8 & mask20))
}

// GetPlmnID extracts PLMNID from the specified ECGI or GEnbID
func GetPlmnID(id uint64) PlmnID {
	return PlmnID(id >> 28)
}

// GetCellID extracts Cell ID from the specified ECI, ECGI or GEnbID
func GetCellID(id uint64) CellID {
	return CellID(id & 0xff)
}

// GetEnbID extracts Enb ID from the specified ECGI or GEnbID
func GetEnbID(id uint64) EnbID {
	return EnbID((id & mask20) >> 8)
}

// GetECI extracts ECI from the specified ECGI or GEnbID
func GetECI(id uint64) ECI {
	return ECI(id & mask28)
}