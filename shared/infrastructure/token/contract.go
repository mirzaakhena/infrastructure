package token

import "github.com/dgrijalva/jwt-go"

type JWTToken interface {
	CreateToken(content interface{}) (string, error)
	VerifyToken(tokenString string) (*jwt.Token, error)
}
