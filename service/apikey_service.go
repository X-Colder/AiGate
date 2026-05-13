package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/aigate/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKeyService struct {
	db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) Create(userID, tenantID string, req *model.CreateAPIKeyRequest) (*model.APIKeyResponse, error) {
	rawKey := generateAPIKey()
	hash := hashKey(rawKey)
	prefix := "sk-..." + rawKey[len(rawKey)-4:]

	qpm := req.RateLimitQPM
	if qpm <= 0 {
		qpm = 60
	}

	apiKey := &model.APIKey{
		ID:           uuid.New().String(),
		UserID:       userID,
		TenantID:     tenantID,
		Name:         req.Name,
		KeyHash:      hash,
		KeyPrefix:    prefix,
		RateLimitQPM: qpm,
		Models:       req.Models,
		Status:       1,
		ExpiresAt:    req.ExpiresAt,
	}
	if err := s.db.Create(apiKey).Error; err != nil {
		return nil, err
	}
	return &model.APIKeyResponse{
		ID:           apiKey.ID,
		Name:         apiKey.Name,
		KeyPrefix:    apiKey.KeyPrefix,
		Key:          rawKey,
		RateLimitQPM: apiKey.RateLimitQPM,
		Models:       apiKey.Models,
		Status:       apiKey.Status,
		ExpiresAt:    apiKey.ExpiresAt,
		CreatedAt:    apiKey.CreatedAt,
	}, nil
}

func (s *APIKeyService) List(userID string) ([]model.APIKeyResponse, error) {
	var keys []model.APIKey
	if err := s.db.Where("user_id = ? AND status = 1", userID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	result := make([]model.APIKeyResponse, len(keys))
	for i, k := range keys {
		result[i] = model.APIKeyResponse{
			ID:           k.ID,
			Name:         k.Name,
			KeyPrefix:    k.KeyPrefix,
			RateLimitQPM: k.RateLimitQPM,
			Models:       k.Models,
			Status:       k.Status,
			LastUsedAt:   k.LastUsedAt,
			ExpiresAt:    k.ExpiresAt,
			CreatedAt:    k.CreatedAt,
		}
	}
	return result, nil
}

func (s *APIKeyService) Revoke(userID, keyID string) error {
	result := s.db.Model(&model.APIKey{}).Where("id = ? AND user_id = ?", keyID, userID).Update("status", 0)
	if result.RowsAffected == 0 {
		return fmt.Errorf("api key not found")
	}
	return result.Error
}

func (s *APIKeyService) ValidateKey(rawKey string) (*model.APIKey, error) {
	hash := hashKey(rawKey)
	var apiKey model.APIKey
	if err := s.db.Where("key_hash = ?", hash).First(&apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invalid api key")
		}
		return nil, err
	}
	if apiKey.Status != 1 {
		return nil, fmt.Errorf("api key disabled")
	}
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("api key expired")
	}
	go s.db.Model(&model.APIKey{}).Where("id = ?", apiKey.ID).Update("last_used_at", time.Now())
	return &apiKey, nil
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "sk-" + hex.EncodeToString(bytes)
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
