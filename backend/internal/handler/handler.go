package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"media-sequencer/backend/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) GetWindows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	windows, err := h.service.GetWindows()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, windows)
}

type AddMediaRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Duration int    `json:"duration"`
}

func (h *Handler) AddMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	windowID, err := getIDFromPath(r.URL.Path, "/windows/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid window id")
		return
	}

	var request AddMediaRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	media, err := h.service.AddMedia(
		windowID,
		request.Name,
		request.Type,
		request.URL,
		request.Duration,
	)

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, media)
}

type SyncRequest struct {
	MediaID  int `json:"media_id"`
	Duration int `json:"duration"`
}

func (h *Handler) SyncMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request SyncRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	syncState, err := h.service.SyncMedia(
		request.MediaID,
		request.Duration,
	)

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, syncState)
}

func (h *Handler) GetSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	syncState, err := h.service.GetSync()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if syncState == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"active": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"active": true,
		"sync":   syncState,
	})
}

func getIDFromPath(path string, prefix string) (int, error) {
	value := strings.TrimPrefix(path, prefix)
	value = strings.Trim(value, "/")

	parts := strings.Split(value, "/")

	if len(parts) == 0 || parts[0] == "" {
		return 0, strconv.ErrSyntax
	}

	return strconv.Atoi(parts[0])
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}