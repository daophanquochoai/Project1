package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Rating struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Rating    int            `gorm:"type:integer;not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   *string        `gorm:"type:text" json:"comment"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (r *Rating) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
