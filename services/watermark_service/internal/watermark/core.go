package watermark

import (
	"fmt"

	"watermark_service/internal/bbox"
)

// Processor applies deterministic, key-based jitter to bounding boxes.
type Processor struct {
	policy NoisePolicy
	prng   *keyedPRNG
}

// NewProcessor builds a Processor with the provided jitter policy using the local secret key.
func NewProcessor(policy NoisePolicy) *Processor {
	prng, err := newKeyedPRNG()
	if err != nil {
		panic(fmt.Errorf("watermark: %w", err))
	}
	return &Processor{
		policy: policy,
		prng:   prng,
	}
}

// Process applies deterministic offsets to each bounding box.
func (p *Processor) Process(boxes []bbox.BBox) []bbox.BBox {
	out := make([]bbox.BBox, len(boxes))
	for i, b := range boxes {
		out[i] = p.policy.Apply(b, i, p.prng)
	}
	return out
}

// VerifyWatermark checks that the provided boxes were produced using the supplied key and policy.
func VerifyWatermark(boxes []bbox.BBox, policy NoisePolicy, key []byte) (bool, error) {
	prng, err := newKeyedPRNGWithKey(key)
	if err != nil {
		return false, err
	}
	for i, b := range boxes {
		if !policy.Matches(b, i, prng) {
			return false, nil
		}
	}
	return true, nil
}
