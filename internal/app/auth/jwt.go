// Package auth provides authentication and authorization services.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"vista-app/internal/app/core"
)

// Claims represents the custom claims embedded in the JWT.
type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTService defines the methods for JWT management.
type JWTService interface {
	Init(cfg core.AuthConfig)
	GenerateToken(userID string, roles []string) (string, error)
	ValidateToken(token string) (*Claims, error)
	RefreshToken(token string) (string, error)
}

// jwtServiceImpl is the concrete implementation of JWTService.
type jwtServiceImpl struct {
	config core.AuthConfig
}

// Common JWT errors
var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token has expired")
	ErrInvalidClaims  = errors.New("invalid token claims")
	ErrTokenNotParsed = errors.New("token could not be parsed")
)

// Global singleton JWT service instance
var (
	jwtService JWTService
)

// Init initializes the JWT service with the provided configuration.
func (j *jwtServiceImpl) Init(cfg core.AuthConfig) {
	j.config = cfg
	jwtService = j
}

// NewJWTService creates a new JWTService instance.
func NewJWTService() JWTService {
	return &jwtServiceImpl{}
}

// GetJWTService returns the global singleton JWT service instance.
func GetJWTService() JWTService {
	return jwtService
}

// GenerateToken creates a new JWT token with the given user info.
func (j *jwtServiceImpl) GenerateToken(
	userID string,
	roles []string,
) (string, error) {
	now := time.Now()
	
	// Parse token expiry duration from config
	expiry, err := time.ParseDuration(j.config.TokenExpiry)
	if err != nil {
		expiry = 1 * time.Hour // Default to 1 hour
	}

	// Create claims with user information
	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
			Subject:   userID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	signedToken, err := token.SignedString(
		[]byte(j.config.JWTSigningKey),
	)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateToken parses and validates a JWT token.
func (j *jwtServiceImpl) ValidateToken(
	tokenString string,
) (*Claims, error) {
	// Parse token with claims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(j.config.JWTSigningKey), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrTokenNotParsed
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Verify issuer
	if claims.Issuer != j.config.Issuer {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshToken generates a new token from an existing valid token.
// It reuses the claims (UserID, Roles) but issues a new expiry time.
func (j *jwtServiceImpl) RefreshToken(
	tokenString string,
) (string, error) {
	// First validate the existing token
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		// If token is expired but otherwise valid, allow refresh
		if !errors.Is(err, ErrExpiredToken) {
			return "", err
		}
		
		// For expired tokens, parse without validation to get claims
		token, parseErr := jwt.ParseWithClaims(
			tokenString,
			&Claims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(j.config.JWTSigningKey), nil
			},
			jwt.WithoutClaimsValidation(),
		)
		
		if parseErr != nil {
			return "", ErrTokenNotParsed
		}
		
		var ok bool
		claims, ok = token.Claims.(*Claims)
		if !ok {
			return "", ErrInvalidClaims
		}
	}

	// Generate new token with same user info
	return j.GenerateToken(claims.UserID, claims.Roles)
}
