package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/BahramRousta/notification-service/internal/notification/service"
)

type Handler struct {
	svc    service.Service
	logger *slog.Logger
}

func New(svc service.Service, l *slog.Logger) *Handler {
	return &Handler{
		svc: svc, logger: l,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/messages/", h.CreateMessage)
}

type CreateMessageRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Recipient      string `json:"recipient"`
	Body           string `json:"body"`
	Priority       int    `json:"priority"`
}

type messageResponse struct {
	ID             string `json:"id"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	Recipient      string `json:"recipient"`
	Body           string `json:"body"`
	Status         string `json:"status"`
	Priority       int    `json:"priority"`
	Attempts       int    `json:"attempts"`
	ProviderMsgID  string `json:"provider_msg_id,omitempty"`
	FailureReason  string `json:"failure_reason,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func toResponse(msg *domain.Message) messageResponse {
	return messageResponse{
		ID:             msg.ID,
		IdempotencyKey: msg.IdempotencyKey,
		Recipient:      msg.Recipient,
		Body:           msg.Body,
		Status:         string(msg.Status),
		Priority:       int(msg.Priority),
		Attempts:       msg.Attempts,
		ProviderMsgID:  msg.ProviderMsgID,
		FailureReason:  msg.FailureReason,
		CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      msg.UpdatedAt.Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, msg)
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req CreateMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	sendReq := service.SendRequest{
		IdempotencyKey: req.IdempotencyKey,
		Recipient:      req.Recipient,
		Body:           req.Body,
		Priority:       domain.Priority(req.Priority),
	}

	msg, err := h.svc.Send(r.Context(), sendReq)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, toResponse(msg))
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRecipient), errors.Is(err, domain.ToLongBodyLen):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrDuplicate):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		h.logger.Error("unexpected service error", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
