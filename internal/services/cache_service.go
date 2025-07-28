package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SAP-2025/auth-service/internal/utils"
	"time"

	"github.com/redis/go-redis/v9"
)

type PKCEStore struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

func NewPKCEStore(client *redis.Client) *PKCEStore {
	return &PKCEStore{
		client: client,
		ctx:    context.Background(),
		ttl:    10 * time.Minute, // PKCE session expires trong 10 phút
	}
}

// Lưu PKCE challenge với session ID
func (s *PKCEStore) SavePKCE(sessionID string, challenge *utils.PKCEChallenge) error {
	key := s.getPKCEKey(sessionID)

	// Serialize PKCE challenge
	data, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("failed to marshal PKCE challenge: %w", err)
	}

	// Save to Redis với TTL
	err = s.client.Set(s.ctx, key, data, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save PKCE to Redis: %w", err)
	}

	return nil
}

// Lấy và xóa PKCE challenge (consume once)
func (s *PKCEStore) GetAndDeletePKCE(sessionID string) (*utils.PKCEChallenge, error) {
	key := s.getPKCEKey(sessionID)

	// Get data
	data, err := s.client.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("PKCE session not found or expired")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get PKCE from Redis: %w", err)
	}

	// Delete immediately (consume once pattern)
	s.client.Del(s.ctx, key)

	// Deserialize
	var challenge utils.PKCEChallenge
	err = json.Unmarshal([]byte(data), &challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal PKCE challenge: %w", err)
	}

	return &challenge, nil
}

// Check if PKCE session exists
func (s *PKCEStore) ExistsPKCE(sessionID string) bool {
	key := s.getPKCEKey(sessionID)
	exists := s.client.Exists(s.ctx, key).Val()
	return exists > 0
}

// Delete PKCE session manually
func (s *PKCEStore) DeletePKCE(sessionID string) error {
	key := s.getPKCEKey(sessionID)
	return s.client.Del(s.ctx, key).Err()
}

// Get TTL of PKCE session
func (s *PKCEStore) GetPKCETTL(sessionID string) (time.Duration, error) {
	key := s.getPKCEKey(sessionID)
	return s.client.TTL(s.ctx, key).Result()
}

// Helper function to generate Redis key
func (s *PKCEStore) getPKCEKey(sessionID string) string {
	return fmt.Sprintf("pkce:session:%s", sessionID)
}

// Clean up expired sessions (optional background job)
func (s *PKCEStore) CleanupExpiredSessions() error {
	pattern := "pkce:session:*"

	iter := s.client.Scan(s.ctx, 0, pattern, 0).Iterator()
	for iter.Next(s.ctx) {
		key := iter.Val()

		// Check TTL
		ttl, err := s.client.TTL(s.ctx, key).Result()
		if err != nil {
			continue
		}

		// If TTL is negative (expired), delete
		if ttl < 0 {
			s.client.Del(s.ctx, key)
		}
	}

	return iter.Err()
}
