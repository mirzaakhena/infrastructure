package token

import (
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const FieldContent = "content"

type jwtToken struct {
	secretKey string
	expired   time.Duration
}

func NewJWTToken(secretKey string, expired time.Duration) (JWTToken, error) {
	if strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("SecretKey must not empty")
	}

	return &jwtToken{
		secretKey: secretKey,
		expired:   expired,
	}, nil
}

func (r *jwtToken) CreateToken(content interface{}) (string, error) {
	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims[FieldContent] = content
	atClaims["exp"] = time.Now().Add(r.expired).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(r.secretKey))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *jwtToken) VerifyToken(tokenString string) (*jwt.Token, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(r.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
