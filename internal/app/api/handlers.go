// Package api provides the HTTP API implementation for the VPN Automation System.
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"vista-app/internal/app/auth"
	"vista-app/internal/app/core"
	"vista-app/internal/app/ipam"
)

// APICallbacks holds handler methods for API endpoints.
type APICallbacks struct {
	cfg     core.Config
	logger  core.LoggerService
	jwtSvc  auth.JWTService
	ipamSvc ipam.Service
}

// Error response structure.
type errorResponse struct {
	Error string `json:"error"`
}

// Login request structure.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login response structure.
type loginResponse struct {
	Token string `json:"token"`
}

// IP reservation request structure.
type reserveIPRequest struct {
	NodeID    string `json:"node_id"`
	PoolID    string `json:"pool_id"`
	Temporary bool   `json:"temporary"`
}

// IP reservation response structure.
type reserveIPResponse struct {
	IP        string `json:"ip"`
	SubnetID  string `json:"subnet_id"`
	LeaseID   string `json:"lease_id"`
	LeaseToken string `json:"lease_token"`
}

// IP commit request structure.
type commitIPRequest struct {
	LeaseToken string            `json:"lease_token"`
	Proofs     map[string]string `json:"proofs"`
}

// IP commit response structure.
type commitIPResponse struct {
	Success bool   `json:"success"`
	LeaseID string `json:"lease_id"`
	IP      string `json:"ip"`
}

// HealthCheck handles GET /api/v1/health requests.
func (a *APICallbacks) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type healthResponse struct {
		Status string `json:"status"`
	}
	respondJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// Login handles POST /api/v1/auth/login requests.
func (a *APICallbacks) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Error("Failed to decode login request", "error", err)
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	token, err := a.jwtSvc.GenerateToken(req.Username, req.Password)
	if err != nil {
		a.logger.Error("Authentication failed", "username", req.Username, "error", err)
		if errors.Is(err, auth.ErrInvalidCredentials) {
			respondError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		respondError(w, http.StatusInternalServerError, "Authentication error")
		return
	}

	a.logger.Info("User authenticated successfully", "username", req.Username)
	respondJSON(w, http.StatusOK, loginResponse{Token: token})
}

// ReserveIP handles POST /api/v1/ipam/reserve-ip requests.
func (a *APICallbacks) ReserveIP(w http.ResponseWriter, r *http.Request) {
	var req reserveIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Error("Failed to decode reserve IP request", "error", err)
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	username := r.Context().Value(contextKeyUsername).(string)

	result, err := a.ipamSvc.ReserveIP(r.Context(), ipam.ReserveRequest{
		NodeID:    req.NodeID,
		PoolID:    req.PoolID,
		Temporary: req.Temporary,
		Username:  username,
	})
	if err != nil {
		a.logger.Error("Failed to reserve IP", 
			"node_id", req.NodeID, 
			"pool_id", req.PoolID, 
			"error", err,
		)
		respondError(w, http.StatusInternalServerError, "Failed to reserve IP")
		return
	}

	a.logger.Info("IP reserved successfully", 
		"ip", result.IP, 
		"lease_id", result.LeaseID,
		"node_id", req.NodeID,
	)
	respondJSON(w, http.StatusOK, reserveIPResponse{
		IP:         result.IP,
		SubnetID:   result.SubnetID,
		LeaseID:    result.LeaseID,
		LeaseToken: result.LeaseToken,
	})
}

// CommitIP handles POST /api/v1/ipam/commit-ip requests (Agent-Claim Model).
func (a *APICallbacks) CommitIP(w http.ResponseWriter, r *http.Request) {
	var req commitIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Error("Failed to decode commit IP request", "error", err)
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	result, err := a.ipamSvc.CommitLease(r.Context(), ipam.CommitRequest{
		LeaseToken: req.LeaseToken,
		Proofs:     req.Proofs,
	})
	if err != nil {
		a.logger.Error("Failed to commit IP lease", "error", err)
		respondError(w, http.StatusInternalServerError, "Failed to commit IP lease")
		return
	}

	a.logger.Info("IP lease committed successfully", 
		"lease_id", result.LeaseID, 
		"ip", result.IP,
	)
	respondJSON(w, http.StatusOK, commitIPResponse{
		Success: true,
		LeaseID: result.LeaseID,
		IP:      result.IP,
	})
}

// RecycleNode handles POST /api/v1/ipam/recycle-node/{nodeID} requests.
func (a *APICallbacks) RecycleNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeID")
	if nodeID == "" {
		a.logger.Error("Missing nodeID in recycle request")
		respondError(w, http.StatusBadRequest, "Missing nodeID parameter")
		return
	}

	username := r.Context().Value(contextKeyUsername).(string)

	count, err := a.ipamSvc.RecycleByNode(r.Context(), nodeID, username)
	if err != nil {
		a.logger.Error("Failed to recycle node resources", "node_id", nodeID, "error", err)
		respondError(w, http.StatusInternalServerError, "Failed to recycle resources")
		return
	}

	type recycleResponse struct {
		Success      bool `json:"success"`
		RecycledCount int  `json:"recycled_count"`
	}

	a.logger.Info("Node resources recycled successfully", 
		"node_id", nodeID, 
		"count", count,
	)
	respondJSON(w, http.StatusOK, recycleResponse{
		Success:      true,
		RecycledCount: count,
	})
}

// respondJSON sends a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log but can't send another response at this point
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// respondError sends a standardized error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, errorResponse{Error: message})
}