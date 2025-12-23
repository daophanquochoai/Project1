package repository

import "errors"

// key redis
var baseUser = "user:"
var baseUserEmail = "user:email:"

// errors
var (
	ErrInternalServer = errors.New("Máy chủ bị lỗi")
	ErrNotFound       = errors.New("Không tìm thấy dữ liệu")
)
