package watermark

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	keyEnvVar       = "WATERMARK_SECRET"
	keyFileEnvVar   = "WATERMARK_KEY_PATH"
	defaultKeyFile  = "watermark.key"
	minKeyByteCount = 32
)

// keyedPRNG deterministically derives pseudo-random scalars in [0,1) using HMAC-SHA256.
type keyedPRNG struct {
	key []byte
}

func newKeyedPRNG() (*keyedPRNG, error) {
	key, err := loadKeyMaterial()
	if err != nil {
		return nil, err
	}
	return &keyedPRNG{key: key}, nil
}

func newKeyedPRNGWithKey(key []byte) (*keyedPRNG, error) {
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	if len(bytes.TrimSpace(keyCopy)) < minKeyByteCount {
		return nil, fmt.Errorf("watermark: key must be at least %d bytes", minKeyByteCount)
	}
	return &keyedPRNG{key: keyCopy}, nil
}

// Scalar returns a deterministic float in [0,1) derived from the key and provided metadata.
func (p *keyedPRNG) Scalar(label string, idx int, dims ...float64) float64 {
	if len(p.key) == 0 {
		return 0
	}

	mac := hmac.New(sha256.New, p.key)
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[:], uint64(idx))
	mac.Write(buf[:])
	mac.Write([]byte(label))

	for _, d := range dims {
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(d))
		mac.Write(buf[:])
	}

	sum := mac.Sum(nil)
	val := binary.BigEndian.Uint64(sum[:8])
	return float64(val) / float64(math.MaxUint64)
}

func loadKeyMaterial() ([]byte, error) {
	if secret := strings.TrimSpace(os.Getenv(keyEnvVar)); secret != "" {
		return normalizeKey([]byte(secret))
	}

	path := strings.TrimSpace(os.Getenv(keyFileEnvVar))
	if path == "" {
		path = defaultKeyFile
	}
	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("watermark: read key file %s: %w", path, err)
	}
	return normalizeKey(data)
}

func normalizeKey(data []byte) ([]byte, error) {
	key := bytes.TrimSpace(data)
	if len(key) < minKeyByteCount {
		return nil, errors.New("watermark: insufficient key length (>=32 bytes required)")
	}
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	return keyCopy, nil
}
