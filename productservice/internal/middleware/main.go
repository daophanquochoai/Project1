package middleware

type Middleware struct {
	Auth *AuthMiddleware
}

func NewMiddleware(auth *AuthMiddleware) *Middleware {
	return &Middleware{Auth: auth}
}
