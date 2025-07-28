package models

import (
	"gorm.io/gorm"
	"net"
)

type AuthLog struct {
	gorm.Model
	UserID       uint
	EventType    string `gorm:"not null"` // login, logout, etc.
	IPAddress    net.IP
	UserAgent    string
	Success      bool `gorm:"default:true"`
	ErrorMessage string
}
