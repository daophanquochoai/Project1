package handler

import "errors"

var (
	ErrInvalidData  = errors.New("Dữ liệu không hợp lệ")
	ErrUnAuthorized = errors.New("Không thể xác thực")
)
