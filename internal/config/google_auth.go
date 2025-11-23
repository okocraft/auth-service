package config

import "os"

type GoogleAuthConfig struct {
	Enabled       bool
	RedirectURL   string
	ClientID      string
	ClientSecret  string
	ResultPageURL string
}

func NewGoogleAuthConfigFromEnv() (GoogleAuthConfig, error) {
	if os.Getenv("AUTH_SERVICE_GOOGLE_AUTH_ENABLED") != "true" {
		return GoogleAuthConfig{}, nil
	}

	redirectURL, err := getRequiredString("AUTH_SERVICE_GOOGLE_AUTH_REDIRECT_URL")
	if err != nil {
		return GoogleAuthConfig{}, err
	}

	clientID, err := getRequiredString("AUTH_SERVICE_GOOGLE_AUTH_CLIENT_ID")
	if err != nil {
		return GoogleAuthConfig{}, err
	}

	clientSecret, err := getRequiredString("AUTH_SERVICE_GOOGLE_AUTH_CLIENT_SECRET")
	if err != nil {
		return GoogleAuthConfig{}, err
	}

	resultPageURL, err := getRequiredString("AUTH_SERVICE_GOOGLE_AUTH_RESULT_PAGE_URL")
	if err != nil {
		return GoogleAuthConfig{}, err
	}

	return GoogleAuthConfig{
		Enabled:       true,
		RedirectURL:   redirectURL,
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		ResultPageURL: resultPageURL,
	}, nil
}
