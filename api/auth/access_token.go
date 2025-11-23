package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	BaseClaims
}

func (c AccessTokenClaims) CreateJWTClaims() jwt.Claims {
	claims := jwt.MapClaims{}

	c.SaveBaseClaimsTo(claims)

	return claims
}

func ReadAccessTokenClaimsFrom(claims jwt.MapClaims) (AccessTokenClaims, error) {
	base, err := ReadBaseClaimsFrom(claims)
	if err != nil {
		return AccessTokenClaims{}, err
	}

	ret := AccessTokenClaims{
		BaseClaims: base,
	}

	return ret, nil
}
