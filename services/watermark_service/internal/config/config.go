package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAddress  = ":8080"
	defaultMinShift = 0.01
	defaultMaxShift = 0.03
	portEnv         = "PORT"
	minShiftPctEnv  = "SHIFT_MIN_PCT"
	maxShiftPctEnv  = "SHIFT_MAX_PCT"
)

// Config describes runtime parameters for the HTTP service.
type Config struct {
	Address       string
	MinShiftRatio float64
	MaxShiftRatio float64
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

	minShift := readFloat(minShiftPctEnv, defaultMinShift)
	maxShift := readFloat(maxShiftPctEnv, defaultMaxShift)
	if minShift > maxShift {
		minShift, maxShift = maxShift, minShift
	}

	return Config{
		Address:       addr,
		MinShiftRatio: minShift,
		MaxShiftRatio: maxShift,
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
