package config

import (
	"os"
	"strconv"
	"time"

	"github.com/Siroshun09/serrors"
)

func getRequiredString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", serrors.New("env '" + key + "' is required")
	}
	return value, nil
}

func getBoolFromEnv(key string, defaultValue bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return defaultValue, nil
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, serrors.WithStackTrace(err)
	}

	return b, nil
}

func getDurationFromEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return defaultValue, nil
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}

	return d, nil
}
