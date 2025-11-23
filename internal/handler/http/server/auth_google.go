package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/Siroshun09/logs"
	"github.com/okocraft/auth-service/api/auth"
	"github.com/okocraft/auth-service/api/user"
	"github.com/okocraft/auth-service/internal/config"
	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/handler/http/oapi"
	"github.com/okocraft/auth-service/internal/usecases"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleAuthHandler struct {
	enabled          bool
	resultPageURL    string
	conf             oauth2.Config
	accessLogUsecase usecases.AccessLogUsecase
	authUsecase      usecases.AuthUsecase
	userUsecase      usecases.UserUsecase
}

func newGoogleAuthHandler(c config.GoogleAuthConfig, accessLogUsecase usecases.AccessLogUsecase, authUsecase usecases.AuthUsecase, userUsecase usecases.UserUsecase) googleAuthHandler {
	return googleAuthHandler{
		enabled:       c.Enabled,
		resultPageURL: c.ResultPageURL,
		conf: oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			RedirectURL:  c.RedirectURL,
			Scopes:       []string{"openid"},
			Endpoint:     google.Endpoint,
		},
		accessLogUsecase: accessLogUsecase,
		authUsecase:      authUsecase,
		userUsecase:      userUsecase,
	}
}

func (h googleAuthHandler) LinkWithGoogle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.enabled {
		httplib.RenderBadRequest(ctx, w, nil)
		return
	}

	req, err := httplib.DecodeJSONRequestBody[oapi.GoogleFirstLoginRequest](r)
	if err != nil {
		httplib.RenderBadRequest(ctx, w, err)
		return
	}

	parsedLoginKey, err := domain.ParseLoginKey(req.LoginKey)
	if err != nil {
		httplib.RenderBadRequest(ctx, w, err)
		return
	}

	verifier := oauth2.GenerateVerifier()
	state, err := h.authUsecase.CreateStateJWTWithLoginKey(ctx, parsedLoginKey, verifier)
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	h.renderGoogleLoginResponse(ctx, w, state, verifier)
}

func (h googleAuthHandler) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.enabled {
		httplib.RenderBadRequest(ctx, w, nil)
		return
	}

	req, err := httplib.DecodeJSONRequestBody[oapi.GoogleLoginRequest](r)
	if err != nil {
		httplib.RenderBadRequest(ctx, w, err)
		return
	}

	verifier := oauth2.GenerateVerifier()

	state, err := h.authUsecase.CreateStateJWT(ctx, req.CurrentUrl, verifier)
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	h.renderGoogleLoginResponse(ctx, w, state, verifier)
}

func (h googleAuthHandler) renderGoogleLoginResponse(ctx context.Context, w http.ResponseWriter, state string, verifier string) {
	redirectURL := h.conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	res, err := httplib.JSONResponse(oapi.GoogleLoginResponse{RedirectUrl: redirectURL})
	if err != nil {
		httplib.RenderInternalServerError(ctx, w, err)
		return
	}

	err = httplib.RenderOKWithBody(ctx, w, res)
	if err != nil {
		logs.Error(ctx, err)
	}
}

func (h googleAuthHandler) CallbackFromGoogle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.enabled {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultNotEnabled)
		return
	}

	state := r.URL.Query().Get("state")

	claimType, claims, err := h.authUsecase.VerifyStateJWT(ctx, state)
	if err != nil {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	}

	var encryptedVerifier string
	var callbackHandleFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, openID string)
	switch claimType {
	case auth.LoginStateClaimTypeUnknown:
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	case auth.LoginStateClaimTypeLogin:
		loginStateClaims, err := auth.ReadLoginStateClaimsFrom(claims)
		if err != nil {
			h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
			return
		}

		encryptedVerifier = loginStateClaims.EncryptedCodeVerifier
		callbackHandleFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request, openID string) {
			h.handleLoginCallback(w, r, openID, loginStateClaims.CurrentPageURL)
		}
	case auth.LoginStateClaimTypeFirstLogin:
		firstLoginStateClaims, err := auth.ReadFirstLoginStateClaimsFrom(claims)
		if err != nil {
			h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
			return
		}

		encryptedVerifier = firstLoginStateClaims.EncryptedCodeVerifier
		callbackHandleFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request, openID string) {
			h.handleFirstLoginCallback(ctx, w, r, openID, domain.LoginKey(firstLoginStateClaims.LoginKey))
		}
	default:
		logs.Errorf(ctx, "unknown claim type: %v", claimType)
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInternalError)
		return
	}

	verifier, err := h.authUsecase.DecryptCodeVerifier(ctx, encryptedVerifier)
	if err != nil {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.conf.Exchange(r.Context(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	}

	jwt := strings.Split(idToken, ".")
	payload := strings.TrimSuffix(jwt[1], "=")
	b, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	}

	var extraClaims map[string]interface{}
	if err := json.Unmarshal(b, &extraClaims); err != nil {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	}

	openID, ok := extraClaims["sub"].(string)
	if !ok {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInvalidToken)
		return
	}

	callbackHandleFunc(ctx, w, r, openID)
}

func (h googleAuthHandler) handleFirstLoginCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, openID string, loginKey domain.LoginKey) {
	usr, err := h.userUsecase.SaveSubByLoginKey(ctx, loginKey, openID)
	switch {
	case errors.Is(err, domain.UserNotFoundByLoginKeyError):
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultLoginKeyNotFound)
		return
	case errors.Is(err, domain.SubAlreadyLinkedError):
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultAlreadyLinked)
		return
	case err != nil:
		logs.Error(ctx, err)
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInternalError)
		return
	}

	h.sendTokens(ctx, w, r, usr, "", domain.AccessLogActionTypeFirstLogin)
}

func (h googleAuthHandler) handleLoginCallback(w http.ResponseWriter, r *http.Request, openID string, redirectTo string) {
	ctx := r.Context()
	usr, err := h.userUsecase.GetUserIDBySub(ctx, openID)
	if errors.Is(err, domain.UserNotFoundBySubError) {
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultUserNotFound)
		return
	} else if err != nil {
		logs.Error(ctx, err)
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInternalError)
		return
	}

	h.sendTokens(ctx, w, r, usr, redirectTo, domain.AccessLogActionTypeLogin)
}

func (h googleAuthHandler) sendTokens(ctx context.Context, w http.ResponseWriter, r *http.Request, userID user.ID, redirectTo string, action domain.AccessLogActionType) {
	loginID, refreshToken, expiresAt, err := h.authUsecase.CreateRefreshToken(ctx, userID)
	if err != nil {
		logs.Error(ctx, err)
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInternalError)
		return
	}

	log := httplib.GetRequestLogFromContext(ctx)
	err = h.accessLogUsecase.SaveAccessLogByUserID(ctx, userID, domain.AccessLogParams{
		Action:    action,
		LoginID:   loginID,
		IP:        log.GetIP(),
		UserAgent: domain.TruncateUserAgent(log.UserAgent),
		CreatedAt: time.Now(),
	})
	if err != nil {
		logs.Error(ctx, err)
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInternalError)
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		logs.Error(ctx, err)
		h.redirectToResultPage(ctx, w, r, oapi.GoogleLoginResultInternalError)
		return
	}

	setRefreshTokenCookie(w, refreshToken, csrfToken, expiresAt)
	httplib.RenderRedirect(ctx, w, r, h.createResultPageURL(oapi.GoogleLoginResultSuccess, redirectTo))
}

func (h googleAuthHandler) redirectToResultPage(ctx context.Context, w http.ResponseWriter, r *http.Request, result oapi.GoogleLoginResult) {
	httplib.RenderRedirect(ctx, w, r, h.createResultPageURL(result, ""))
}

func (h googleAuthHandler) createResultPageURL(result oapi.GoogleLoginResult, redirectTo string) string {
	redirect := h.resultPageURL + "?type=" + string(result)
	if redirectTo != "" {
		redirect += "&redirectTo=" + url.PathEscape(redirectTo)
	}
	return redirect
}
