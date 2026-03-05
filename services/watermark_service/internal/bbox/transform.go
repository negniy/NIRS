package bbox

import "fmt"

// BBox represents a rectangular bounding box in absolute coordinates.
type BBox struct {
	XMin float64
	YMin float64
	XMax float64
	YMax float64
}

// Width returns the box width (may be negative if coordinates are invalid).
func (b BBox) Width() float64 {
	return b.XMax - b.XMin
}

// Height returns the box height (may be negative if coordinates are invalid).
func (b BBox) Height() float64 {
	return b.YMax - b.YMin
}

// Shift returns a new bounding box translated by the provided deltas.
func (b BBox) Shift(dx, dy float64) BBox {
	return BBox{
		XMin: b.XMin + dx,
		YMin: b.YMin + dy,
		XMax: b.XMax + dx,
		YMax: b.YMax + dy,
	}
}

// FromSlice converts a slice [xMin, yMin, xMax, yMax] into a BBox.
func FromSlice(vals []float64) (BBox, error) {
	if len(vals) != 4 {
		return BBox{}, fmt.Errorf("bbox slice must have length 4, got %d", len(vals))
	}
	return BBox{
		XMin: vals[0],
		YMin: vals[1],
		XMax: vals[2],
		YMax: vals[3],
	}, nil
}

// ToSlice converts the bounding box into an array convenient for JSON encoding.
func (b BBox) ToSlice() [4]float64 {
	return [4]float64{b.XMin, b.YMin, b.XMax, b.YMax}
}
