package dto

type ServiceResponse struct {
	Err    error
	Status int
}

type PageResponse struct {
	Total  int64       `json:"total"`
	Data   interface{} `json:"data"`
	Filter any         `json:"filter"`
}
