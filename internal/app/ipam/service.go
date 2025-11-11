// Package ipam provides IP and Port Address Management services.
package ipam

import (
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"vista-app/internal/app/core"
)

// IPAMService defines operational methods for IPAM/PortAM module.
type IPAMService interface {
	Init(db *gorm.DB, cfg core.IPAMConfig)
	ReserveIP(nodeID, tunnelID string, dur time.Duration) (
		*IPLease, error)
	CommitIP(token, ip string) error
	ReleaseIPByTunnel(tunnelID string) error
	ReservePort(nodeID, proto, tunnelID string, dur time.Duration) (
		*PortLease, error)
	CommitPort(token string, port int) error
	ReleasePortByTunnel(tunnelID string) error
	RecycleExpiredLeases() (int, error)
	RecycleByNode(nodeID string) (int, error)
}

// ipamServiceImpl is the concrete implementation.
type ipamServiceImpl struct {
	db     *gorm.DB
	config core.IPAMConfig
	logger core.LoggerService
}

var (
	ErrNoAvailableIP   = errors.New("no available IP in pool")
	ErrNoAvailablePort = errors.New("no available port in range")
	ErrLeaseNotFound   = errors.New("lease not found")
	ipamService        IPAMService
)

func (s *ipamServiceImpl) Init(db *gorm.DB, cfg core.IPAMConfig) {
	s.db, s.config, s.logger = db, cfg, core.Log()
	ipamService = s
}

func NewIPAMService() IPAMService { return &ipamServiceImpl{} }
func GetIPAMService() IPAMService { return ipamService }

func (s *ipamServiceImpl) ReserveIP(
	nodeID, tunnelID string, dur time.Duration,
) (*IPLease, error) {
	var lease *IPLease
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("node_id = ? AND state = ?", nodeID,
			StateAvailable).First(&lease).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoAvailableIP
			}
			return err
		}
		now, exp := time.Now(), time.Now().Add(dur)
		lease.State, lease.UserID = StateReserved, &tunnelID
		lease.ReservedAt, lease.ExpiresAt = &now, &exp
		if err := tx.Save(lease).Error; err != nil {
			return err
		}
		return s.audit(tx, "IPLease", lease.ID, "reserve",
			StateAvailable, StateReserved, &tunnelID, "IP reserved")
	}); err != nil {
		s.logger.Error("Reserve IP failed",
			zap.String("node_id", nodeID), zap.Error(err))
		return nil, err
	}
	s.logger.Info("IP reserved", zap.String("ip", lease.IP))
	return lease, nil
}

func (s *ipamServiceImpl) CommitIP(token, ip string) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var lease IPLease
		if err := tx.Where("user_id = ? AND ip = ? AND state = ?",
			token, ip, StateReserved).First(&lease).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrLeaseNotFound
			}
			return err
		}
		now := time.Now()
		lease.State, lease.AllocatedAt, lease.ExpiresAt = StateUsed, &now, nil
		if err := tx.Save(&lease).Error; err != nil {
			return err
		}
		return s.audit(tx, "IPLease", lease.ID, "commit",
			StateReserved, StateUsed, &token, "IP committed")
	}); err != nil {
		s.logger.Error("Commit IP failed", zap.String("ip", ip),
			zap.Error(err))
		return err
	}
	s.logger.Info("IP committed", zap.String("ip", ip))
	return nil
}

func (s *ipamServiceImpl) ReleaseIPByTunnel(tunnelID string) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var leases []IPLease
		if err := tx.Where("user_id = ?", tunnelID).Find(&leases).
			Error; err != nil {
			return err
		}
		for _, l := range leases {
			old := l.State
			transitionToAvailable(&l)
			if err := tx.Save(&l).Error; err != nil {
				return err
			}
			if err := s.audit(tx, "IPLease", l.ID, "release",
				old, StateAvailable, &tunnelID,
				"IP released"); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		s.logger.Error("Release IPs failed",
			zap.String("tunnel_id", tunnelID), zap.Error(err))
		return err
	}
	s.logger.Info("IPs released", zap.String("tunnel_id", tunnelID))
	return nil
}

func (s *ipamServiceImpl) ReservePort(
	nodeID, proto, tunnelID string, dur time.Duration,
) (*PortLease, error) {
	var lease *PortLease
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var usedPorts []int
		var node NodeCapability
		if err := tx.Where("node_id = ?", nodeID).First(&node).
			Error; err == nil && node.UsedPorts != "" {
			_ = json.Unmarshal([]byte(node.UsedPorts), &usedPorts)
		}
		q := tx.Where("node_id = ? AND protocol = ? AND state = ?",
			nodeID, proto, StateAvailable)
		if len(usedPorts) > 0 {
			q = q.Where("port NOT IN ?", usedPorts)
		}
		if err := q.First(&lease).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoAvailablePort
			}
			return err
		}
		now, exp := time.Now(), time.Now().Add(dur)
		lease.State, lease.UserID = StateReserved, &tunnelID
		lease.ReservedAt, lease.ExpiresAt = &now, &exp
		if err := tx.Save(lease).Error; err != nil {
			return err
		}
		return s.audit(tx, "PortLease", lease.ID, "reserve",
			StateAvailable, StateReserved, &tunnelID, "Port reserved")
	}); err != nil {
		s.logger.Error("Reserve port failed",
			zap.String("node_id", nodeID), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Port reserved", zap.Int("port", lease.Port))
	return lease, nil
}

func (s *ipamServiceImpl) CommitPort(token string, port int) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var lease PortLease
		if err := tx.Where("user_id = ? AND port = ? AND state = ?",
			token, port, StateReserved).First(&lease).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrLeaseNotFound
			}
			return err
		}
		now := time.Now()
		lease.State, lease.AllocatedAt, lease.ExpiresAt = StateUsed, &now, nil
		if err := tx.Save(&lease).Error; err != nil {
			return err
		}
		return s.audit(tx, "PortLease", lease.ID, "commit",
			StateReserved, StateUsed, &token, "Port committed")
	}); err != nil {
		s.logger.Error("Commit port failed", zap.Int("port", port),
			zap.Error(err))
		return err
	}
	s.logger.Info("Port committed", zap.Int("port", port))
	return nil
}

func (s *ipamServiceImpl) ReleasePortByTunnel(tunnelID string) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var leases []PortLease
		if err := tx.Where("user_id = ?", tunnelID).Find(&leases).
			Error; err != nil {
			return err
		}
		for _, l := range leases {
			old := l.State
			transitionToAvailable(&l)
			if err := tx.Save(&l).Error; err != nil {
				return err
			}
			if err := s.audit(tx, "PortLease", l.ID, "release",
				old, StateAvailable, &tunnelID,
				"Port released"); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		s.logger.Error("Release ports failed",
			zap.String("tunnel_id", tunnelID), zap.Error(err))
		return err
	}
	s.logger.Info("Ports released",
		zap.String("tunnel_id", tunnelID))
	return nil
}

func (s *ipamServiceImpl) RecycleExpiredLeases() (int, error) {
	recycled, now := 0, time.Now()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ipCount, err := s.recycleIPLeases(tx,
			"state = ? AND expires_at < ?", StateReserved, now)
		if err != nil {
			return err
		}
		portCount, err := s.recyclePortLeases(tx,
			"state = ? AND expires_at < ?", StateReserved, now)
		if err != nil {
			return err
		}
		recycled = ipCount + portCount
		return nil
	})
	if err != nil {
		s.logger.Error("Recycle expired leases failed",
			zap.Error(err))
		return 0, err
	}
	s.logger.Info("Expired leases recycled",
		zap.Int("count", recycled))
	return recycled, nil
}

func (s *ipamServiceImpl) RecycleByNode(nodeID string) (int, error) {
	recycled := 0
	states := []LeaseState{StateReserved, StateUsed}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ipCount, err := s.recycleIPLeases(tx,
			"node_id = ? AND state IN ?", nodeID, states)
		if err != nil {
			return err
		}
		portCount, err := s.recyclePortLeases(tx,
			"node_id = ? AND state IN ?", nodeID, states)
		if err != nil {
			return err
		}
		recycled = ipCount + portCount
		return nil
	})
	if err != nil {
		s.logger.Error("Recycle node resources failed",
			zap.String("node_id", nodeID), zap.Error(err))
		return 0, err
	}
	s.logger.Info("Node resources recycled",
		zap.String("node_id", nodeID), zap.Int("count", recycled))
	return recycled, nil
}
