package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"templespace/cmd/auth/internal/rbac"
	"templespace/cmd/auth/internal/storage"
	"time"
)

type MagicStore interface {
	Save(token string, email string, ttl time.Duration)
	Get(token string) (string, bool)
	Delete(token string)
}

type RefreshStore interface {
	Save(token string, userID string, ttl time.Duration)
	Get(token string) (string, bool)
	Delete(token string)
}

type Service struct {
	Signer       *JWTSigner
	MagicTTL     time.Duration
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	MagicStore   MagicStore
	RefreshStore RefreshStore
	Users        storage.UserRepo
}

func NewService(signer *JWTSigner, magicTTL, accessTTL, refreshTTL time.Duration, ms MagicStore, rs RefreshStore) *Service {
	return &Service{Signer: signer, MagicTTL: magicTTL, AccessTTL: accessTTL, RefreshTTL: refreshTTL, MagicStore: ms, RefreshStore: rs}
}

func (s *Service) generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) StartLogin(email string) (string, error) {
	token, err := s.generateToken(16)
	if err != nil {
		return "", err
	}
	s.MagicStore.Save(token, email, s.MagicTTL)
	return token, nil
}

func (s *Service) VerifyMagicToken(token string) (access string, refresh string, expSec int64, err error) {
	email, ok := s.MagicStore.Get(token)
	if !ok {
		return "", "", 0, errors.New("invalid or expired magic token")
	}
	s.MagicStore.Delete(token)
	// Lookup/create user and derive scopes by role (default: user)
	role := "user"
	userID := email
	if s.Users != nil {
		if u, ok := s.Users.FindByEmail(email); ok {
			userID = u.ID
			if u.Role != "" {
				role = u.Role
			}
		} else {
			_ = s.Users.Upsert(storage.User{ID: email, Email: email, Role: role})
		}
	}
	scopes := rbac.ScopesFor(rbac.Role(role))
	claims := Claims{Subject: userID, UserID: userID, Email: email, Scopes: scopes, Expires: time.Now().Add(s.AccessTTL).Unix()}
	access, err = s.Signer.Sign(claims)
	if err != nil {
		return "", "", 0, err
	}
	refresh, err = s.generateToken(16)
	if err != nil {
		return "", "", 0, err
	}
	s.RefreshStore.Save(refresh, userID, s.RefreshTTL)
	return access, refresh, int64(s.AccessTTL.Seconds()), nil
}
