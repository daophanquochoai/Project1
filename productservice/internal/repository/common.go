package repository

import "errors"

var (
	baseProduct        = "product:"
	baseProductSimilar = "productsimilar:"
	baseProductRelated = "productrelated:"
	baseRateOfUser     = "rateofuser:"
	baseRateOfProduct  = "rateofproduct:"
	baseRateStatistic  = "ratestatistic:"
)

var (
	ErrInternalServerError = errors.New("Máy chủ bị lỗi")
	ErrNotFound            = errors.New("Dữ liệu không tìm thấy")
	ErrWasRateProduct      = errors.New("Sản phẩm đã được đánh giá")
	ErrForbidden           = errors.New("Bạn không có quyền thực hiện")
)
