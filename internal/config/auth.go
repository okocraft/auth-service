package config

import (
	"encoding/hex"
	"time"

	"github.com/Siroshun09/serrors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/okocraft/authlib/encrypt"
	"github.com/okocraft/authlib/jwtclaims"
)

type AuthConfig struct {
	Encrypter                  encrypt.Encrypter
	JWTSigner                  jwtclaims.JWTSigner
	LoginExpireDuration        time.Duration
	AccessTokenExpireDuration  time.Duration
	RefreshTokenExpireDuration time.Duration
}

func NewAuthConfigFromEnv() (AuthConfig, error) {
	privateKeyHex, err := getRequiredString("AUTH_SERVICE_PRIVATE_KEY")
	if err != nil {
		return AuthConfig{}, err
	}

	privateKey, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return AuthConfig{}, serrors.Errorf("failed to decode AUTH_SERVICE_PRIVATE_KEY: %w", err)
	} else if len(privateKey) != 32 {
		return AuthConfig{}, serrors.New("AUTH_SERVICE_PRIVATE_KEY must be hex value of 32 bytes long")
	}

	encrypter, err := encrypt.NewAESEncrypter(privateKey)
	if err != nil {
		return AuthConfig{}, serrors.Errorf("failed to create encrypter: %w", err)
	}

	jwtSigner := jwtclaims.NewJWTSigner(jwt.SigningMethodHS512, privateKey)

	loginExpire, err := getDurationFromEnv("AUTH_SERVICE_LOGIN_EXPIRE", 15*time.Minute)
	if err != nil {
		return AuthConfig{}, err
	}

	accessTokenExpire, err := getDurationFromEnv("AUTH_SERVICE_ACCESS_TOKEN_EXPIRE", 15*time.Minute)
	if err != nil {
		return AuthConfig{}, err
	}

	refreshTokenExpire, err := getDurationFromEnv("AUTH_SERVICE_REFRESH_TOKEN_EXPIRE", 7*24*time.Hour)
	if err != nil {
		return AuthConfig{}, err
	}

	return AuthConfig{
		Encrypter:                  encrypter,
		JWTSigner:                  jwtSigner,
		LoginExpireDuration:        loginExpire,
		AccessTokenExpireDuration:  accessTokenExpire,
		RefreshTokenExpireDuration: refreshTokenExpire,
	}, nil
}
