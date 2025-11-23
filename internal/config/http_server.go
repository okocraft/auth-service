package config

import (
	"os"
	"strings"
)

type HTTPServerConfig struct {
	Debug            bool
	Port             string
	AllowedOrigins   map[string]struct{}
	DBConfig         DBConfig
	AuthConfig       AuthConfig
	GoogleAuthConfig GoogleAuthConfig
}

func NewHTTPServerConfigFromEnv() (HTTPServerConfig, error) {
	debug, err := getBoolFromEnv("DEBUG", false)
	if err != nil {
		return HTTPServerConfig{}, err
	}

	port, err := getRequiredString("AUTH_SERVICE_HTTP_PORT")
	if err != nil {
		return HTTPServerConfig{}, err
	}

	origins := createOriginSet(os.Getenv("AUTH_SERVICE_ALLOWED_ORIGINS"))

	dbConfig, err := NewDBConfigFromEnv()
	if err != nil {
		return HTTPServerConfig{}, err
	}

	authConfig, err := NewAuthConfigFromEnv()
	if err != nil {
		return HTTPServerConfig{}, err
	}

	googleAuthConfig, err := NewGoogleAuthConfigFromEnv()
	if err != nil {
		return HTTPServerConfig{}, err
	}

	return HTTPServerConfig{
		Debug:            debug,
		Port:             port,
		AllowedOrigins:   origins,
		DBConfig:         dbConfig,
		AuthConfig:       authConfig,
		GoogleAuthConfig: googleAuthConfig,
	}, nil
}

func createOriginSet(value string) map[string]struct{} {
	if value == "" {
		return map[string]struct{}{}
	}

	origins := strings.Split(value, ",")
	if len(origins) == 0 {
		return map[string]struct{}{}
	}

	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		set[origin] = struct{}{}
	}
	return set
}
