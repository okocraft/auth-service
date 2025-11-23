package auth

import (
	"github.com/Siroshun09/serrors"
	"github.com/golang-jwt/jwt/v5"
)

type JWTSigner interface {
	Sign(claims jwt.Claims) (string, error)
	VerifyAndParse(token string) (jwt.MapClaims, error)
}

func NewJWTSigner(method jwt.SigningMethod, hmacSecret []byte) JWTSigner {
	return &jwtSigner{method: method, hmacSecret: hmacSecret}
}

type jwtSigner struct {
	method     jwt.SigningMethod
	hmacSecret []byte
}

func (s jwtSigner) Sign(claims jwt.Claims) (string, error) {
	tokenString, err := jwt.NewWithClaims(s.method, claims).SignedString(s.hmacSecret)
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}
	return tokenString, nil
}

func (s jwtSigner) VerifyAndParse(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, serrors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.hmacSecret, nil
	})
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, serrors.New("invalid claims")
}
