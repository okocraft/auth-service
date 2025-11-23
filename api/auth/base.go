package auth

import (
	"fmt"
	"time"

	"github.com/Siroshun09/serrors"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type BaseClaims struct {
	JTI       uuid.UUID
	NotBefore time.Time
	ExpiresAt time.Time
}

func (c BaseClaims) SaveBaseClaimsTo(claims jwt.MapClaims) {
	claims["jti"] = c.JTI.String()
	claims["exp"] = jwt.NewNumericDate(c.ExpiresAt)
	claims["nbf"] = jwt.NewNumericDate(c.NotBefore)
}

func (c BaseClaims) Validate(now time.Time) error {
	if c.JTI.IsNil() {
		return serrors.New("missing jti claim")
	}

	if c.NotBefore.Before(now) {
		return serrors.Errorf("not before %s, but now is %s", c.NotBefore, now)
	}

	if c.ExpiresAt.Equal(now) || c.ExpiresAt.Before(now) {
		return serrors.Errorf("token is expired at %s", c.ExpiresAt)
	}

	return nil
}

func ReadBaseClaimsFrom(claims jwt.MapClaims) (BaseClaims, error) {
	jti, err := uuid.FromString(fmt.Sprintf("%v", claims["jti"]))
	if err != nil {
		return BaseClaims{}, serrors.WithStackTrace(err)
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return BaseClaims{}, serrors.WithStackTrace(err)
	} else if exp == nil {
		return BaseClaims{}, serrors.New("missing exp claim")
	}

	nbf, err := claims.GetNotBefore()
	if err != nil {
		return BaseClaims{}, serrors.WithStackTrace(err)
	} else if nbf == nil {
		return BaseClaims{}, serrors.New("missing nbf claim")
	}

	return BaseClaims{
		JTI:       jti,
		ExpiresAt: exp.Time,
		NotBefore: nbf.Time,
	}, nil
}
