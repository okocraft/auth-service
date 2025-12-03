package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/okocraft/auth-service/internal/config"
	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/repositories"
	"github.com/okocraft/auth-service/internal/repositories/database"
	"github.com/okocraft/authlib/jwtclaims"
	"github.com/okocraft/authlib/user"

	"github.com/Siroshun09/serrors"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUsecase interface {
	CreateStateJWT(ctx context.Context, currentPageURL string, codeVerifier string) (string, error)
	CreateStateJWTWithLoginKey(ctx context.Context, loginKey domain.LoginKey, codeVerifier string) (string, error)
	VerifyStateJWT(ctx context.Context, tokenString string) (jwtclaims.LoginStateClaimType, jwt.MapClaims, error)
	GetUserIDAndRefreshTokenIDFromJTI(ctx context.Context, jti uuid.UUID) (user.ID, int64, error)
	DecryptCodeVerifier(ctx context.Context, encryptedCodeVerifier string) (string, error)
	CreateRefreshToken(ctx context.Context, userID user.ID) (uuid.UUID, string, time.Time, error)
	VerifyRefreshToken(ctx context.Context, tokenString string) (jwtclaims.RefreshTokenClaims, error)
	RefreshToken(ctx context.Context, params domain.RefreshTokenParams) (domain.RefreshedToken, error)
	InvalidateTokens(ctx context.Context, refreshTokenClaims jwtclaims.RefreshTokenClaims) error
	CreateLoginKey(ctx context.Context, userID user.ID) (domain.LoginKey, error)
}

func NewAuthUsecase(conf config.AuthConfig, db database.DB, repo repositories.AuthRepository, userRepo repositories.UserRepository) AuthUsecase {
	return authUsecase{
		conf:     conf,
		db:       db,
		repo:     repo,
		userRepo: userRepo,
	}
}

type authUsecase struct {
	conf     config.AuthConfig
	db       database.DB
	repo     repositories.AuthRepository
	userRepo repositories.UserRepository
}

func (u authUsecase) CreateStateJWT(_ context.Context, currentPageURL string, codeVerifier string) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	encryptedCodeVerifier, err := u.conf.Encrypter.Encrypt([]byte(codeVerifier))
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	createdAt := time.Now()
	expiresAt := createdAt.Add(u.conf.LoginExpireDuration)

	state := jwtclaims.LoginStateClaims{
		BaseClaims: jwtclaims.BaseClaims{
			JTI:       id,
			NotBefore: time.Now(),
			ExpiresAt: expiresAt,
		},
		CurrentPageURL:        currentPageURL,
		EncryptedCodeVerifier: hex.EncodeToString(encryptedCodeVerifier),
	}

	tokenString, err := u.conf.JWTSigner.Sign(state.CreateJWTClaims())
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	return tokenString, nil
}

func (u authUsecase) CreateStateJWTWithLoginKey(_ context.Context, loginKey domain.LoginKey, codeVerifier string) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	encryptedCodeVerifier, err := u.conf.Encrypter.Encrypt([]byte(codeVerifier))
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	createdAt := time.Now()
	expiresAt := createdAt.Add(u.conf.LoginExpireDuration)

	state := jwtclaims.FirstLoginStateClaims{
		BaseClaims: jwtclaims.BaseClaims{
			JTI:       id,
			NotBefore: time.Now(),
			ExpiresAt: expiresAt,
		},
		LoginKey:              int64(loginKey),
		EncryptedCodeVerifier: hex.EncodeToString(encryptedCodeVerifier),
	}

	tokenString, err := u.conf.JWTSigner.Sign(state.CreateJWTClaims())
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	return tokenString, nil
}

func (u authUsecase) VerifyStateJWT(_ context.Context, tokenString string) (jwtclaims.LoginStateClaimType, jwt.MapClaims, error) {
	claims, err := u.conf.JWTSigner.VerifyAndParse(tokenString)
	if err != nil {
		return jwtclaims.LoginStateClaimTypeUnknown, nil, serrors.WithStackTrace(err)
	}

	claimType := jwtclaims.GetLoginStateClaimType(claims)
	return claimType, claims, nil
}

func (u authUsecase) GetUserIDAndRefreshTokenIDFromJTI(ctx context.Context, jti uuid.UUID) (user.ID, int64, error) {
	userID, refreshTokenID, err := u.repo.GetUserIDAndRefreshTokenIDFromJTI(ctx, u.db.Conn(), jti)
	if errors.Is(err, domain.RefreshTokenIDByJTINotFoundError) {
		return 0, 0, serrors.WithStackTrace(domain.NewUnauthorizedError(err))
	} else if err != nil {
		return 0, 0, serrors.WithStackTrace(err)
	}
	return userID, refreshTokenID, nil
}

func (u authUsecase) DecryptCodeVerifier(_ context.Context, encryptedCodeVerifier string) (string, error) {
	data, err := hex.DecodeString(encryptedCodeVerifier)
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	decrypted, err := u.conf.Encrypter.Decrypt(data)
	if err != nil {
		return "", serrors.WithStackTrace(err)
	}

	return string(decrypted), nil
}

func (u authUsecase) CreateRefreshToken(ctx context.Context, userID user.ID) (uuid.UUID, string, time.Time, error) {
	refreshTokenJTI, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, "", time.Time{}, serrors.WithStackTrace(err)
	}

	loginID, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, "", time.Time{}, serrors.WithStackTrace(err)
	}

	createdAt := time.Now()
	expiresAt := createdAt.Add(u.conf.RefreshTokenExpireDuration)

	err = u.repo.SaveRefreshToken(ctx, u.db.Conn(), userID, refreshTokenJTI, loginID, expiresAt)
	if err != nil {
		return uuid.Nil, "", time.Time{}, serrors.WithStackTrace(err)
	}

	refreshToken := jwtclaims.RefreshTokenClaims{
		BaseClaims: jwtclaims.BaseClaims{
			JTI:       refreshTokenJTI,
			NotBefore: createdAt,
			ExpiresAt: expiresAt,
		},
		LoginID: loginID,
	}

	refreshTokenString, err := u.conf.JWTSigner.Sign(refreshToken.CreateJWTClaims())
	if err != nil {
		return uuid.Nil, "", time.Time{}, serrors.WithStackTrace(err)
	}

	return loginID, refreshTokenString, createdAt, nil
}

func (u authUsecase) VerifyRefreshToken(_ context.Context, tokenString string) (jwtclaims.RefreshTokenClaims, error) {
	claims, err := u.conf.JWTSigner.VerifyAndParse(tokenString)
	if err != nil {
		return jwtclaims.RefreshTokenClaims{}, serrors.WithStackTrace(err)
	}

	refreshTokenClaims, err := jwtclaims.ReadRefreshTokenClaimsFrom(claims)
	if err != nil {
		return jwtclaims.RefreshTokenClaims{}, serrors.WithStackTrace(err)
	}

	if err := refreshTokenClaims.Validate(time.Now()); err != nil {
		return jwtclaims.RefreshTokenClaims{}, serrors.WithStackTrace(err)
	}

	return refreshTokenClaims, nil
}

func (u authUsecase) RefreshToken(ctx context.Context, params domain.RefreshTokenParams) (domain.RefreshedToken, error) {
	refreshTokenJTI, err := uuid.NewV7()
	if err != nil {
		return domain.RefreshedToken{}, serrors.WithStackTrace(err)
	}

	accessTokenJTI, err := uuid.NewV7()
	if err != nil {
		return domain.RefreshedToken{}, serrors.WithStackTrace(err)
	}

	createdAt := time.Now()

	expiresAt := createdAt.Add(u.conf.AccessTokenExpireDuration)
	if expiresAt.After(params.MaxExpiresAt) {
		expiresAt = params.MaxExpiresAt
	}

	err = u.db.WithTx(ctx, func(ctx context.Context, tx database.Connection) error {
		err = u.repo.SaveAccessToken(ctx, tx, params.RefreshTokenID, accessTokenJTI, createdAt)
		if err != nil {
			return serrors.WithStackTrace(err)
		}

		err = u.repo.SaveRefreshToken(ctx, tx, params.UserID, refreshTokenJTI, params.LoginID, createdAt)
		if err != nil {
			return serrors.WithStackTrace(err)
		}

		return nil
	})
	if err != nil {
		return domain.RefreshedToken{}, serrors.WithStackTrace(err)
	}

	refreshToken := jwtclaims.RefreshTokenClaims{
		BaseClaims: jwtclaims.BaseClaims{
			JTI:       refreshTokenJTI,
			NotBefore: createdAt,
			ExpiresAt: expiresAt,
		},
		LoginID: params.LoginID,
	}

	accessToken := jwtclaims.AccessTokenClaims{
		BaseClaims: jwtclaims.BaseClaims{
			JTI:       accessTokenJTI,
			NotBefore: createdAt,
			ExpiresAt: expiresAt,
		},
	}

	refreshTokenString, err := u.conf.JWTSigner.Sign(refreshToken.CreateJWTClaims())
	if err != nil {
		return domain.RefreshedToken{}, serrors.WithStackTrace(err)
	}

	accessTokenString, err := u.conf.JWTSigner.Sign(accessToken.CreateJWTClaims())
	if err != nil {
		return domain.RefreshedToken{}, serrors.WithStackTrace(err)
	}

	return domain.RefreshedToken{
		RefreshToken: refreshTokenString,
		AccessToken:  accessTokenString,
		ExpiresAt:    expiresAt,
	}, nil
}

func (u authUsecase) InvalidateTokens(ctx context.Context, refreshTokenClaims jwtclaims.RefreshTokenClaims) error {
	err := u.db.WithTx(ctx, func(ctx context.Context, tx database.Connection) error {
		err := u.repo.DeleteAccessTokensByLoginID(ctx, tx, refreshTokenClaims.LoginID)
		if err != nil {
			return serrors.WithStackTrace(err)
		}

		err = u.repo.DeleteRefreshTokensByLoginID(ctx, tx, refreshTokenClaims.LoginID)
		if err != nil {
			return serrors.WithStackTrace(err)
		}

		return nil
	})
	if err != nil {
		return serrors.WithStackTrace(err)
	}

	return nil
}

func (u authUsecase) CreateLoginKey(ctx context.Context, userID user.ID) (domain.LoginKey, error) {
	key, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return 0, serrors.WithStackTrace(err)
	}

	loginKey := domain.LoginKey(key.Int64())

	err = u.userRepo.SaveLoginKeyForUserID(ctx, u.db.Conn(), userID, loginKey, time.Now())
	if err != nil {
		return 0, serrors.WithStackTrace(err)
	}

	return loginKey, nil
}
