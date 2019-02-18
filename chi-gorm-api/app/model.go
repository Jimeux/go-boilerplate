package app

import (
	"time"
)

type Model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	Name      string     `json:"name" validate:"required"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

func (Model) TableName() string {
	return "model"
}
