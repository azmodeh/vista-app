// Package ipam provides IP and Port Address Management models.
package ipam

import (
	"time"

	"gorm.io/gorm"
)

// LeaseState represents the current state of an IP or Port resource.
type LeaseState string

const (
	StateAvailable LeaseState = "Available"
	StateReserved  LeaseState = "Reserved"
	StateUsed      LeaseState = "Used"
)

// IPPool defines the pool boundaries and governance rules for IPs.
type IPPool struct {
	gorm.Model
	CIDR        string `gorm:"type:varchar(50);uniqueIndex;not null"`
	NodeID      string `gorm:"type:varchar(100);index"`
	IsGlobal    bool   `gorm:"default:false;index"`
	Description string `gorm:"type:text"`
}

// IPLease represents the core model for IP allocation (the 'Loan').
type IPLease struct {
	gorm.Model
	IP          string      `gorm:"type:varchar(45);uniqueIndex;not null"`
	PoolID      uint        `gorm:"index;not null"`
	Pool        IPPool      `gorm:"foreignKey:PoolID"`
	NodeID      string      `gorm:"type:varchar(100);index"`
	State       LeaseState  `gorm:"type:varchar(20);index;not null"`
	UserID      *string     `gorm:"type:varchar(100);index"`
	ReservedAt  *time.Time  `gorm:"index"`
	AllocatedAt *time.Time  `gorm:"index"`
	ExpiresAt   *time.Time  `gorm:"index"`
	Metadata    string      `gorm:"type:jsonb"`
}

// PortPool defines the port range boundaries and protocol.
type PortPool struct {
	gorm.Model
	RangeStart  int    `gorm:"not null;index"`
	RangeEnd    int    `gorm:"not null;index"`
	Protocol    string `gorm:"type:varchar(10);index;not null"`
	NodeID      string `gorm:"type:varchar(100);index"`
	Blacklist   string `gorm:"type:jsonb"`
	Description string `gorm:"type:text"`
}

// PortLease represents the core model for Port allocation (the 'Loan').
type PortLease struct {
	gorm.Model
	Port        int         `gorm:"index;not null"`
	Protocol    string      `gorm:"type:varchar(10);index;not null"`
	PoolID      uint        `gorm:"index;not null"`
	Pool        PortPool    `gorm:"foreignKey:PoolID"`
	NodeID      string      `gorm:"type:varchar(100);index"`
	State       LeaseState  `gorm:"type:varchar(20);index;not null"`
	UserID      *string     `gorm:"type:varchar(100);index"`
	ReservedAt  *time.Time  `gorm:"index"`
	AllocatedAt *time.Time  `gorm:"index"`
	ExpiresAt   *time.Time  `gorm:"index"`
	Metadata    string      `gorm:"type:jsonb"`
}

// Audit tracks resource recycling (Bailiff action) and IPAM changes.
type Audit struct {
	gorm.Model
	ResourceType string     `gorm:"type:varchar(20);index;not null"`
	ResourceID   uint       `gorm:"index;not null"`
	Action       string     `gorm:"type:varchar(50);index;not null"`
	OldState     LeaseState `gorm:"type:varchar(20)"`
	NewState     LeaseState `gorm:"type:varchar(20)"`
	UserID       *string    `gorm:"type:varchar(100);index"`
	Reason       string     `gorm:"type:text"`
	Metadata     string     `gorm:"type:jsonb"`
}

// NodeCapability stores node capabilities and heartbeat information.
type NodeCapability struct {
	gorm.Model
	NodeID         string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	Capabilities   string     `gorm:"type:jsonb"`
	UsedPorts      string     `gorm:"type:jsonb"`
	LastHeartbeat  time.Time  `gorm:"index"`
	Status         string     `gorm:"type:varchar(20);index"`
	IPAddress      string     `gorm:"type:varchar(45);index"`
	Version        string     `gorm:"type:varchar(50)"`
	Metadata       string     `gorm:"type:jsonb"`
}

// TableName overrides the table name for IPPool.
func (IPPool) TableName() string {
	return "ip_pools"
}

// TableName overrides the table name for IPLease.
func (IPLease) TableName() string {
	return "ip_leases"
}

// TableName overrides the table name for PortPool.
func (PortPool) TableName() string {
	return "port_pools"
}

// TableName overrides the table name for PortLease.
func (PortLease) TableName() string {
	return "port_leases"
}

// TableName overrides the table name for Audit.
func (Audit) TableName() string {
	return "audits"
}

// TableName overrides the table name for NodeCapability.
func (NodeCapability) TableName() string {
	return "node_capabilities"
}

// IsAvailable checks if the lease state is Available.
func (ls LeaseState) IsAvailable() bool {
	return ls == StateAvailable
}

// IsReserved checks if the lease state is Reserved.
func (ls LeaseState) IsReserved() bool {
	return ls == StateReserved
}

// IsUsed checks if the lease state is Used.
func (ls LeaseState) IsUsed() bool {
	return ls == StateUsed
}

// String returns the string representation of LeaseState.
func (ls LeaseState) String() string {
	return string(ls)
}

// IsExpired checks if the IP lease has expired.
func (l *IPLease) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

// CanTransitionTo validates if state transition is allowed.
func (l *IPLease) CanTransitionTo(newState LeaseState) bool {
	switch l.State {
	case StateAvailable:
		return newState == StateReserved
	case StateReserved:
		return newState == StateUsed || newState == StateAvailable
	case StateUsed:
		return newState == StateAvailable
	default:
		return false
	}
}

// IsExpired checks if the port lease has expired.
func (l *PortLease) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

// CanTransitionTo validates if state transition is allowed.
func (l *PortLease) CanTransitionTo(newState LeaseState) bool {
	switch l.State {
	case StateAvailable:
		return newState == StateReserved
	case StateReserved:
		return newState == StateUsed || newState == StateAvailable
	case StateUsed:
		return newState == StateAvailable
	default:
		return false
	}
}

// IsOnline checks if the node is considered online.
// A node is online if heartbeat is within the last 5 minutes.
func (n *NodeCapability) IsOnline(threshold time.Duration) bool {
	if threshold == 0 {
		threshold = 5 * time.Minute // Default threshold
	}
	return time.Since(n.LastHeartbeat) <= threshold
}
