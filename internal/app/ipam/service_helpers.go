// Package ipam provides IP and Port Address Management services.
package ipam

import "gorm.io/gorm"

// Helper: create audit record
func (s *ipamServiceImpl) audit(
	tx *gorm.DB, resType string, resID uint, action string,
	old, new LeaseState, userID *string, reason string,
) error {
	return tx.Create(&Audit{
		ResourceType: resType, ResourceID: resID, Action: action,
		OldState: old, NewState: new, UserID: userID, Reason: reason,
	}).Error
}

// Helper: transition lease to available
func transitionToAvailable(lease interface{}) {
	switch l := lease.(type) {
	case *IPLease:
		l.State, l.UserID = StateAvailable, nil
		l.ReservedAt, l.AllocatedAt, l.ExpiresAt = nil, nil, nil
	case *PortLease:
		l.State, l.UserID = StateAvailable, nil
		l.ReservedAt, l.AllocatedAt, l.ExpiresAt = nil, nil, nil
	}
}

// Helper: recycle IP leases
func (s *ipamServiceImpl) recycleIPLeases(
	tx *gorm.DB, query string, args ...interface{},
) (int, error) {
	var leases []IPLease
	if err := tx.Where(query, args...).Find(&leases).Error; err != nil {
		return 0, err
	}
	count := 0
	for _, lease := range leases {
		oldState, userID := lease.State, lease.UserID
		transitionToAvailable(&lease)
		if err := tx.Save(&lease).Error; err != nil {
			return count, err
		}
		if err := s.audit(tx, "IPLease", lease.ID, "recycle",
			oldState, StateAvailable, userID,
			"Lease recycled by Bailiff"); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// Helper: recycle port leases
func (s *ipamServiceImpl) recyclePortLeases(
	tx *gorm.DB, query string, args ...interface{},
) (int, error) {
	var leases []PortLease
	if err := tx.Where(query, args...).Find(&leases).Error; err != nil {
		return 0, err
	}
	count := 0
	for _, lease := range leases {
		oldState, userID := lease.State, lease.UserID
		transitionToAvailable(&lease)
		if err := tx.Save(&lease).Error; err != nil {
			return count, err
		}
		if err := s.audit(tx, "PortLease", lease.ID, "recycle",
			oldState, StateAvailable, userID,
			"Lease recycled by Bailiff"); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
