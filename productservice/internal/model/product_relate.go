package model

import (
	"github.com/google/uuid"
	"time"
)

type ProductRelate struct {
	ProductID    uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"product_id"`
	RelatedID    uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"related_id"`
	RelationType string    `gorm:"type:text;not null;check:relation_type IN ('related', 'similar')" json:"relation_type"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`

	// Relations
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Related Product `gorm:"foreignKey:RelatedID" json:"related,omitempty"`
}
