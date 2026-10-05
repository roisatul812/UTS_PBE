package model

type ApiResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    interface{}         `json:"data,omitempty"`
	Meta    *PaginationMeta     `json:"meta,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}
