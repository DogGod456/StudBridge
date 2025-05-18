package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"sbChat/internal/usecase/chat"
)

var (
	ErrInvalidRequestMethod = errors.New("only POST method is allowed")
	ErrInvalidRequestBody   = errors.New("failed to parse request body")
	ErrResponseEncoding     = errors.New("failed to encode response payload")
)

type CreateChatHandler struct {
	createChatUseCase *chat.CreateChatUseCase
}

func NewCreateChatHandler(createChatUseCase *chat.CreateChatUseCase) *CreateChatHandler {
	return &CreateChatHandler{
		createChatUseCase: createChatUseCase,
	}
}

type ParticipantInput struct {
	ID   string `json:"id"` // ref_ID
	Type string `json:"type"`
}

type CreateChatRequest struct {
	Participants []ParticipantInput `json:"participants"`
}

type CreateChatResponse struct {
	ChatID string `json:"chat_id"`
}

func (h *CreateChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("%s: %v", ErrInvalidRequestMethod, http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	var req CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("%w: invalid JSON format", ErrInvalidRequestBody),
			http.StatusBadRequest)
		return
	}

	if len(req.Participants) == 0 {
		http.Error(w, fmt.Sprintf("validation failed: %w", chat.ErrEmptyParticipants),
			http.StatusBadRequest)
		return
	}

	// Конвертируем локальные ParticipantInput в тип из usecase
	ucParticipants := make([]chat.ParticipantInput, len(req.Participants))
	for i, p := range req.Participants {
		ucParticipants[i] = chat.ParticipantInput{
			ID:   p.ID,
			Type: p.Type,
		}
	}

	chatID, err := h.createChatUseCase.Execute(r.Context(), ucParticipants)
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrEmptyParticipants):
			http.Error(w, fmt.Sprintf("validation error: %v", err),
				http.StatusBadRequest)
		case errors.Is(err, chat.ErrInvalidParticipantData):
			http.Error(w, fmt.Sprintf("invalid input: %v", err),
				http.StatusUnprocessableEntity)
		default:
			http.Error(w, "internal server error: failed to process chat creation",
				http.StatusInternalServerError)
		}
		return
	}

	resp := CreateChatResponse{
		ChatID: chatID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("%w: %v", ErrResponseEncoding, err),
			http.StatusInternalServerError)
	}
}
