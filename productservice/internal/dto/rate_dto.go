package dto

import (
	"time"

	"github.com/google/uuid"
)

type MyRatingsRequest struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	SortBy string `query:"sort"`
}

type MyRatingsResponse struct {
	Total      int64            `json:"total"`
	Data       []*RateResponse  `json:"data"`
	Pagination any              `json:"pagination"`
	Summary    *MyRatingSummary `json:"summary"`
}

type MyRatingSummary struct {
	TotalRatings  int64   `json:"total_ratings"`
	AvgStarsGiven float64 `json:"avg_stars_given"`
}

type RatingSummaryResponse struct {
	ProductID         string                 `json:"product_id"`
	Summary           *RatingSummary         `json:"summary"`
	Distribution      map[string]*StarDetail `json:"distribution"`
	DistributionChart []*DistributionChart   `json:"distribution_chart"`
}

type RatingSummary struct {
	AvgRating    float64 `json:"avg_rating"`
	TotalRatings int64   `json:"total_ratings"`
}

type StarDetail struct {
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type DistributionChart struct {
	Stars      int     `json:"stars"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type RateListOfProduct struct {
	ProductId uuid.UUID `json:"product_id"`
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
	Stars     int       `json:"stars"`
	SortBy    string    `json:"sort"`
}

type UpdateRatingProduct struct {
	RateId  uuid.UUID `json:"rate_id"`
	Star    int       `json:"stars"`
	Comment string    `json:"comment"`
}

type RateProduct struct {
	UserId    uuid.UUID
	ProductId uuid.UUID
	Star      int    `json:"stars"`
	Comment   string `json:"comment"`
}

type RateResponse struct {
	Id       uuid.UUID     `json:"id"`
	User     UserOfRate    `json:"userId,omitempty"`
	Product  ProductOfRate `json:"product"`
	Star     int           `json:"stars"`
	Comment  string        `json:"comment"`
	CreateAt time.Time     `json:"createAt,omitempty"`
	UpdateAt time.Time     `json:"updateAt,omitempty"`
}

type UserOfRate struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ProductOfRate struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	AverageRating float64   `json:"averageRating"`
	TotalRatings  int       `json:"totalRatings"`
}
