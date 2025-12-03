package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/Siroshun09/logs"
	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/handler/http/oapi"
	"github.com/okocraft/auth-service/internal/usecases"
)

type authHandler struct {
	authUsecase      usecases.AuthUsecase
	accessLogUsecase usecases.AccessLogUsecase
}

func newAuthHandler(authUsecase usecases.AuthUsecase, accessLogUsecase usecases.AccessLogUsecase) authHandler {
	return authHandler{
		authUsecase:      authUsecase,
		accessLogUsecase: accessLogUsecase,
	}
}

func (h authHandler) Logout(w http.ResponseWriter, r *http.Request, params oapi.LogoutParams) {
	ctx := r.Context()

	if err := checkCSRFToken(r, params.XCSRFToken); err != nil {
		httplib.RenderNoContentForUnauthorized(ctx, w, err)
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httplib.RenderNoContentForUnauthorized(ctx, w, err)
		return
	}

	unsetRefreshTokenCookie(w)

	refreshTokenClaims, err := h.authUsecase.VerifyRefreshToken(ctx, cookie.Value)
	if domain.IsUnauthorizedError(err) {
		httplib.RenderNoContentForUnauthorized(ctx, w, err) // already logged out
		return
	} else if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	userID, _, err := h.authUsecase.GetUserIDAndRefreshTokenIDFromJTI(ctx, refreshTokenClaims.JTI)
	if domain.IsUnauthorizedError(err) {
		httplib.RenderNoContentForUnauthorized(ctx, w, err)
		return
	} else if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	err = h.authUsecase.InvalidateTokens(ctx, refreshTokenClaims)
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	log := httplib.GetRequestLogFromContext(ctx)
	err = h.accessLogUsecase.SaveAccessLogByUserID(ctx, userID, domain.AccessLogParams{
		Action:    domain.AccessLogActionTypeLogout,
		LoginID:   refreshTokenClaims.LoginID,
		IP:        log.GetIP(),
		UserAgent: domain.TruncateUserAgent(log.UserAgent),
		CreatedAt: time.Now(),
	})
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	httplib.RenderNoContent(ctx, w)
}

func (h authHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request, params oapi.RefreshAccessTokenParams) {
	ctx := r.Context()

	if err := checkCSRFToken(r, params.XCSRFToken); err != nil {
		httplib.RenderUnauthorized(ctx, w, err)
		return
	}

	refreshTokenClaims, err := h.authUsecase.VerifyRefreshToken(ctx, params.RefreshToken)
	if err != nil {
		httplib.RenderUnauthorized(ctx, w, err)
		return
	}

	userID, refreshTokenID, err := h.authUsecase.GetUserIDAndRefreshTokenIDFromJTI(ctx, refreshTokenClaims.JTI)
	if errors.Is(err, domain.RefreshTokenIDByJTINotFoundError) {
		httplib.RenderUnauthorized(ctx, w, err)
		return
	} else if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	token, err := h.authUsecase.RefreshToken(ctx, domain.RefreshTokenParams{
		UserID:         userID,
		RefreshTokenID: refreshTokenID,
		LoginID:        refreshTokenClaims.LoginID,
		MaxExpiresAt:   refreshTokenClaims.ExpiresAt,
	})
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	log := httplib.GetRequestLogFromContext(ctx)
	err = h.accessLogUsecase.SaveAccessLogByUserID(ctx, userID, domain.AccessLogParams{
		Action:    domain.AccessLogActionTypeRefreshToken,
		LoginID:   refreshTokenClaims.LoginID,
		IP:        log.GetIP(),
		UserAgent: domain.TruncateUserAgent(log.UserAgent),
		CreatedAt: time.Now(),
	})
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	setRefreshTokenCookie(w, token.RefreshToken, csrfToken, token.ExpiresAt)

	res, err := httplib.JSONResponse(oapi.AccessTokenResponse{
		AccessToken: token.AccessToken,
	})
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	err = httplib.RenderOKWithBody(ctx, w, res)
	if err != nil {
		logs.Error(ctx, err)
	}
}
