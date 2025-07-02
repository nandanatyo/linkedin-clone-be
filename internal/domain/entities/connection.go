package entities

import (
	"gorm.io/gorm"
	"time"
)

type ConnectionStatus string

const (
	ConnectionPending  ConnectionStatus = "pending"
	ConnectionAccepted ConnectionStatus = "accepted"
	ConnectionBlocked  ConnectionStatus = "blocked"
)

type Connection struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	RequesterID uint             `gorm:"not null" json:"requester_id"`
	AddresseeID uint             `gorm:"not null" json:"addressee_id"`
	Status      ConnectionStatus `gorm:"default:'pending'" json:"status"`
	RequestedAt time.Time        `json:"requested_at"`
	AcceptedAt  *time.Time       `json:"accepted_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"index" json:"-"`

	Requester User `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	Addressee User `gorm:"foreignKey:AddresseeID" json:"addressee,omitempty"`
}

func (Connection) TableName() string {
	return "connections"
}

type ConnectionIndex struct {
	RequesterID uint `gorm:"index:idx_connection_requester_addressee,unique"`
	AddresseeID uint `gorm:"index:idx_connection_requester_addressee,unique"`
}
