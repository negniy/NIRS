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

// NoisePolicy controls the deterministic hybrid watermark configuration.
type NoisePolicy struct {
	Alpha   float64
	Beta    float64
	Enabled bool
}

// Apply deterministically expands (optionally) and jitters a bounding box according to the keyed PRNG.
func (p NoisePolicy) Apply(box bbox.BBox, idx int, prng *keyedPRNG) bbox.BBox {
	if !p.Enabled || prng == nil {
		return box
	}

	width := box.Width()
	height := box.Height()
	if width <= 0 || height <= 0 {
		return box
	}

	cx := (box.XMin + box.XMax) * 0.5
	cy := (box.YMin + box.YMax) * 0.5

	scale := structuralScale(idx, p.Beta)
	scaledWidth := width * scale
	scaledHeight := height * scale

	dx, dy := p.centerShift(width, height, idx, prng)
	cx += dx
	cy += dy

	return reconstructBox(cx, cy, scaledWidth, scaledHeight)
}

// Matches reports whether the provided bounding box conforms to the deterministic watermark.
func (p NoisePolicy) Matches(box bbox.BBox, idx int, prng *keyedPRNG) bool {
	if !p.Enabled || prng == nil {
		return true
	}

	width := box.Width()
	height := box.Height()
	if width <= 0 || height <= 0 {
		return false
	}

	scale := structuralScale(idx, p.Beta)
	if scale == 0 {
		return false
	}

	dx, dy := p.centerShift(width, height, idx, prng)

	cx := (box.XMin + box.XMax) * 0.5
	cy := (box.YMin + box.YMax) * 0.5

	baseCx := cx - dx
	baseCy := cy - dy
	baseWidth := width / scale
	baseHeight := height / scale

	neutral := reconstructBox(baseCx, baseCy, baseWidth, baseHeight)
	expected := p.Apply(neutral, idx, prng)

	return boxesClose(expected, box)
}

func (p NoisePolicy) centerShift(width, height float64, idx int, prng *keyedPRNG) (float64, float64) {
	alpha := math.Abs(p.Alpha)
	if alpha == 0 || width <= 0 || height <= 0 {
		return 0, 0
	}

	rx := prng.Scalar(axisX, idx, width, height)
	ry := prng.Scalar(axisY, idx, width, height)

	sx := 2*rx - 1
	sy := 2*ry - 1

	return alpha * width * sx, alpha * height * sy
}

func structuralScale(idx int, beta float64) float64 {
	if !isTriggerObject(idx) {
		return 1
	}
	scale := 1 + beta
	if scale <= 0 {
		return 1
	}
	return scale
}

func reconstructBox(cx, cy, width, height float64) bbox.BBox {
	halfW := width * 0.5
	halfH := height * 0.5
	return bbox.BBox{
		XMin: cx - halfW,
		YMin: cy - halfH,
		XMax: cx + halfW,
		YMax: cy + halfH,
	}
}

func boxesClose(a, b bbox.BBox) bool {
	return nearlyEqual(a.XMin, b.XMin) &&
		nearlyEqual(a.XMax, b.XMax) &&
		nearlyEqual(a.YMin, b.YMin) &&
		nearlyEqual(a.YMax, b.YMax)
}

func nearlyEqual(a, b float64) bool {
	diff := math.Abs(a - b)
	scale := math.Max(math.Max(math.Abs(a), math.Abs(b))*toleranceScale, toleranceScale)
	return diff <= scale
}

// isTriggerObject identifies whether a bounding box should receive the structural watermark.
// TODO: implement real trigger detection logic.
func isTriggerObject(idx int) bool {
	return false
}
