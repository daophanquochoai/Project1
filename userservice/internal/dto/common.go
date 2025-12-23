package dto

import models "github.com/agris/user-service/internal/model"

type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    any    `json:"data"`
}

type PageResponse struct {
	Total  int64       `json:"total"`
	Data   interface{} `json:"data"`
	Filter interface{} `json:"filter"`
}

type PageRequest struct {
	Page   int          `json:"page"`
	Limit  int          `json:"limit"`
	Search string       `json:"search"`
	Role   *models.Role `json:"role"`
}

type ServiceResponse struct {
	Err    error
	Status int
}
