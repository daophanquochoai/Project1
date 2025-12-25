package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Product struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name          string         `gorm:"type:text;not null" json:"name"`
	SearchName    string         `gorm:"type:text;not null" json:"search_name"`
	Description   *string        `gorm:"type:text" json:"description"`
	Price         float64        `gorm:"type:numeric(12,2);not null" json:"price"`
	CategoryID    uuid.UUID      `gorm:"type:uuid;not null" json:"category_id"`
	AverageRating float64        `gorm:"type:numeric(3,2);not null;default:0" json:"average_rating"`
	TotalRatings  int            `gorm:"type:integer;not null;default:0" json:"total_ratings"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Category    Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Ratings     []Rating        `gorm:"foreignKey:ProductID" json:"ratings,omitempty"`
	RelatedFrom []ProductRelate `gorm:"foreignKey:ProductID" json:"related_from,omitempty"`
	RelatedTo   []ProductRelate `gorm:"foreignKey:RelatedID" json:"related_to,omitempty"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
