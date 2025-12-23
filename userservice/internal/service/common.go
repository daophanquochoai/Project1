package service

// error
const (
	ErrInvalidData          = "Dữ liệu không khớp"
	ErrUserNotFound         = "Tài khoản không thể tìm thấy"
	ErrEmailExists          = "email already exists"
	ErrHashPassword         = "failed to hash password"
	ErrPasswordNotMatch     = "password not match"
	ErrTokenGenerate        = "failed to generate token"
	ErrLockedAccount        = "Tài khoản đã bị khóa"
	ErrCantUpdateOwnAccount = "Không thể cập nhật tài khoản chính mình"
	ErrInternalServerError  = "Máy chủ bị lỗi"
)
