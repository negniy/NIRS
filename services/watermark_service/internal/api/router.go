package api

import "net/http"

// NewRouter wires HTTP endpoints to the provided handler set.
func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/process-bboxes", h.ProcessBBoxes)
	return mux
}
