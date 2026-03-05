package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"watermark_service/internal/bbox"
)

// BBoxProcessor defines the behaviour required by the HTTP handler.
type BBoxProcessor interface {
	Process([]bbox.BBox) []bbox.BBox
}

// Handler bundles HTTP handlers for the service.
type Handler struct {
	processor BBoxProcessor
}

// NewHandler constructs a Handler instance.
func NewHandler(processor BBoxProcessor) *Handler {
	return &Handler{processor: processor}
}

// ProcessBBoxes handles POST /process-bboxes requests.
func (h *Handler) ProcessBBoxes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload [][]float64
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	boxes := make([]bbox.BBox, 0, len(payload))
	for i, raw := range payload {
		box, err := bbox.FromSlice(raw)
		if err != nil {
			http.Error(w, annotateIndex(err, i), http.StatusBadRequest)
			return
		}
		boxes = append(boxes, box)
	}

	processed := h.processor.Process(boxes)
	response := make([][4]float64, len(processed))
	for i, b := range processed {
		response[i] = b.ToSlice()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("encode error: %v", err), http.StatusInternalServerError)
		return
	}
}

func annotateIndex(err error, idx int) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("bbox[%d]: %v", idx, err)
}
