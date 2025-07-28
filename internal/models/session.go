package models

import (
	"github.com/google/uuid"
	"net"
	"time"
)

type UserSession struct {
	ID               uuid.UUID `gorm:"primaryKey;default:gen_random_uuid()"`
	UserID           uint      `gorm:"references:users(id);onDelete:CASCADE"`
	RefreshTokenHash string    `gorm:"not null"`
	ExpiresAt        time.Time `gorm:"not null"`
	CreatedAt        time.Time `gorm:"default:NOW()"`
	LastUsedAt       time.Time `gorm:"default:NOW()"`
	UserAgent        string
	IPAddress        net.IP
}
