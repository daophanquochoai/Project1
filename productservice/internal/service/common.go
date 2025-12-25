package service

import "errors"

var (
	ErrInvalid  = errors.New("Giá trị lọc không hợp lệ")
	ErrNotFound = errors.New("Không tìm thấy dữ liệu")
)
