package dto

import (
	"github.com/google/uuid"
)

type PageProdRequest struct {
	Page        int         `json:"page"`
	Limit       int         `json:"limit"`
	Search      string      `json:"search"`
	SortBy      string      `json:"sort"`
	SortOrder   string      `json:"order"`
	CategoryIds []uuid.UUID `json:"category_ids"`
	MinPrice    int         `json:"min_price"`
	MaxPrice    int         `json:"max_price"`
	MinRate     int         `json:"min_rate"`
	MaxRate     int         `json:"max_rate"`
}

type PageSimilarAndRelatedRequest struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

type RelationType string

var (
	Relation_Similar RelationType = "similar"
	Relation_Related RelationType = "related"
)
