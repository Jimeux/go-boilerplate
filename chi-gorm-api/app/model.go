package app

import (
	"github.com/jinzhu/gorm"
)

type Model struct {
	// ID   int64  `json:"id"`
	gorm.Model
	Name string `json:"name"`
}

func (Model) TableName() string {
	return "model"
}
