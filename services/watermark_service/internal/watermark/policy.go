package watermark

import (
	"math"

	"watermark_service/internal/bbox"
)

const (
	axisX          = "axis_x"
	axisY          = "axis_y"
	toleranceScale = 1e-6
)

// NoisePolicy controls the magnitude of jitter relative to the box size.
type NoisePolicy struct {
	MinShiftRatio float64
	MaxShiftRatio float64
}

// Apply deterministically shifts a bounding box according to the keyed PRNG.
func (p NoisePolicy) Apply(box bbox.BBox, idx int, prng *keyedPRNG) bbox.BBox {
	width := box.Width()
	height := box.Height()
	if width <= 0 || height <= 0 || prng == nil {
		return box
	}

	dx := p.axisDelta(box.XMin, width, height, idx, axisX, prng)
	dy := p.axisDelta(box.YMin, height, width, idx, axisY, prng)
	return box.Shift(dx, dy)
}

func (p NoisePolicy) axisDelta(anchor, size, companion float64, idx int, axis string, prng *keyedPRNG) float64 {
	minRatio, maxRatio := p.bounds()
	if size <= 0 || maxRatio == 0 {
		return 0
	}

	span := size * maxRatio
	grid := span * 2
	if grid == 0 {
		return 0
	}

	scalar := prng.Scalar(axis, idx, size, companion)
	residual := (scalar - 0.5) * 2 * span
	residual = enforceMinimumMagnitude(residual, size*minRatio)

	base := math.Round(anchor/grid) * grid
	target := base + residual
	return target - anchor
}

// Matches reports whether the provided bounding box conforms to the deterministic watermark.
func (p NoisePolicy) Matches(box bbox.BBox, idx int, prng *keyedPRNG) bool {
	width := box.Width()
	height := box.Height()
	if width <= 0 || height <= 0 || prng == nil {
		return false
	}
	return p.axisMatches(box.XMin, width, height, idx, axisX, prng) &&
		p.axisMatches(box.YMin, height, width, idx, axisY, prng)
}

func (p NoisePolicy) axisMatches(anchor, size, companion float64, idx int, axis string, prng *keyedPRNG) bool {
	minRatio, maxRatio := p.bounds()
	if size <= 0 || maxRatio == 0 {
		return true
	}
	span := size * maxRatio
	grid := span * 2
	if grid == 0 {
		return true
	}

	scalar := prng.Scalar(axis, idx, size, companion)
	expected := (scalar - 0.5) * 2 * span
	expected = enforceMinimumMagnitude(expected, size*minRatio)

	base := math.Round(anchor/grid) * grid
	actual := anchor - base
	tol := math.Max(size*toleranceScale, toleranceScale)
	return math.Abs(actual-expected) <= tol
}

func (p NoisePolicy) bounds() (float64, float64) {
	min := math.Abs(p.MinShiftRatio)
	max := math.Abs(p.MaxShiftRatio)
	if max == 0 {
		max = min
	}
	if min == 0 {
		min = max
	}
	if min > max {
		min, max = max, min
	}
	return min, max
}

func enforceMinimumMagnitude(val, minAbs float64) float64 {
	minAbs = math.Abs(minAbs)
	if minAbs == 0 {
		return val
	}
	if math.Abs(val) < minAbs {
		if val == 0 {
			return minAbs
		}
		return math.Copysign(minAbs, val)
	}
	return val
}
