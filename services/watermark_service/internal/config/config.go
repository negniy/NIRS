package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAddress = ":8080"
	defaultAlpha   = 0.01
	defaultBeta    = 0.05

	portEnv   = "PORT"
	alphaEnv  = "WATERMARK_ALPHA"
	betaEnv   = "WATERMARK_BETA"
	enableEnv = "WATERMARK_ENABLE"
)

// Config describes runtime parameters for the HTTP service.
type Config struct {
	Address         string
	WatermarkAlpha  float64
	WatermarkBeta   float64
	WatermarkEnable bool
}

// Load reads configuration from environment variables and falls back to defaults.
func Load() Config {
	addr := defaultAddress
	if env := strings.TrimSpace(os.Getenv(portEnv)); env != "" {
		if strings.HasPrefix(env, ":") {
			addr = env
		} else {
			addr = fmt.Sprintf(":%s", env)
		}
	}

	alpha := readFloat(alphaEnv, defaultAlpha)
	beta := readFloat(betaEnv, defaultBeta)
	enabled := readBool(enableEnv, true)

	return Config{
		Address:         addr,
		WatermarkAlpha:  alpha,
		WatermarkBeta:   beta,
		WatermarkEnable: enabled,
	}
}

func readFloat(env string, def float64) float64 {
	val := strings.TrimSpace(os.Getenv(env))
	if val == "" {
		return def
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}
	return f
}

func readBool(env string, def bool) bool {
	val := strings.TrimSpace(os.Getenv(env))
	if val == "" {
		return def
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}
	return b
}
