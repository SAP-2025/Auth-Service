package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	CasdoorUserID string `gorm:"unique;not null"`
	Username      string `gorm:"unique;not null"`
	Email         string `gorm:"unique;not null"`
	Name          string `gorm:"not null"`
	AvatarURL     string
	Role          string `gorm:"check:role IN ('student', 'teacher', 'admin', 'proctor')"`
	Organization  string
	IsActive      bool `gorm:"default:true"`
	LastLoginAt   time.Time
}
