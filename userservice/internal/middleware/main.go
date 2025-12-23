package middleware

import "github.com/agris/user-service/pkg/jwtMg"

type Middleware struct {
	Auth *AuthMiddleware
}

func NewMiddleware(jwtManager *jwtMg.JWTManager) *Middleware {
	return &Middleware{Auth: NewAuthMiddleware(jwtManager)}
}
