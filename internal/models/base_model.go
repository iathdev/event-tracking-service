package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID uuid.UUID `gorm:"primaryKey;column:id;type:uuid" json:"id"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.GenerateID()
	}
	return
}

func (b *BaseModel) GenerateID() {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return
	}
	b.ID = uuidV7
}
