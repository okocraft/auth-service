package auth

import (
	"strconv"

	"github.com/Siroshun09/serrors"
	"github.com/golang-jwt/jwt/v5"
)

type LoginStateClaimType uint8

const (
	LoginStateClaimTypeUnknown LoginStateClaimType = iota
	LoginStateClaimTypeLogin
	LoginStateClaimTypeFirstLogin
)

func GetLoginStateClaimType(claims jwt.MapClaims) LoginStateClaimType {
	if _, ok := claims["current_page"]; ok {
		return LoginStateClaimTypeLogin
	}
	if _, ok := claims["login_key"]; ok {
		return LoginStateClaimTypeFirstLogin
	}
	return LoginStateClaimTypeUnknown
}

type LoginStateClaims struct {
	BaseClaims

	CurrentPageURL        string
	EncryptedCodeVerifier string
}

func (c LoginStateClaims) CreateJWTClaims() jwt.Claims {
	claims := jwt.MapClaims{}

	c.BaseClaims.SaveBaseClaimsTo(claims)
	claims["current_page"] = c.CurrentPageURL
	claims["code_verifier"] = c.EncryptedCodeVerifier

	return claims
}

func ReadLoginStateClaimsFrom(claims jwt.MapClaims) (LoginStateClaims, error) {
	base, err := ReadBaseClaimsFrom(claims)
	if err != nil {
		return LoginStateClaims{}, err
	}

	ret := LoginStateClaims{
		BaseClaims: base,
	}

	currentPageURL, ok := claims["current_page"].(string)
	if ok { // optional
		ret.CurrentPageURL = currentPageURL
	}

	codeVerifier, ok := claims["code_verifier"].(string)
	if !ok {
		return LoginStateClaims{}, serrors.New("missing code_verifier claim")
	}
	ret.EncryptedCodeVerifier = codeVerifier

	return ret, nil
}

type FirstLoginStateClaims struct {
	BaseClaims

	LoginKey              int64
	EncryptedCodeVerifier string
}

func (c FirstLoginStateClaims) CreateJWTClaims() jwt.Claims {
	claims := jwt.MapClaims{}

	c.BaseClaims.SaveBaseClaimsTo(claims)
	claims["login_key"] = strconv.FormatInt(c.LoginKey, 16)
	claims["code_verifier"] = c.EncryptedCodeVerifier

	return claims
}

func ReadFirstLoginStateClaimsFrom(claims jwt.MapClaims) (FirstLoginStateClaims, error) {
	base, err := ReadBaseClaimsFrom(claims)
	if err != nil {
		return FirstLoginStateClaims{}, err
	}

	ret := FirstLoginStateClaims{
		BaseClaims: base,
	}

	rawLoginKey, ok := claims["login_key"].(string)
	if !ok {
		return FirstLoginStateClaims{}, serrors.New("missing login_key claim")
	}

	loginKey, err := strconv.ParseInt(rawLoginKey, 16, 64)
	if err != nil {
		return FirstLoginStateClaims{}, serrors.WithStackTrace(err)
	}
	ret.LoginKey = loginKey

	codeVerifier, ok := claims["code_verifier"].(string)
	if !ok {
		return FirstLoginStateClaims{}, serrors.New("missing code_verifier claim")
	}
	ret.EncryptedCodeVerifier = codeVerifier

	return ret, nil
}
