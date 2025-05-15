package chat

import (
	"encoding/json"
	"net/http"

	"sbChat/internal/usecase/chat"
)

type CreateChatHandler struct {
	createChatUseCase *chat.CreateChatUseCase
}

func NewCreateChatHandler(createChatUseCase *chat.CreateChatUseCase) *CreateChatHandler {
	return &CreateChatHandler{
		createChatUseCase: createChatUseCase,
	}
}

type CreateChatRequest struct {
	ParticipantIDs []string `json:"participant_ids"`
}

type CreateChatResponse struct {
	ChatID string `json:"chat_id"`
}

func (h *CreateChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.ParticipantIDs) == 0 {
		http.Error(w, "at least one participant is required", http.StatusBadRequest)
		return
	}

	chatID, err := h.createChatUseCase.Execute(r.Context(), req.ParticipantIDs...)
	if err != nil {
		switch err {
		case chat.ErrChatAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		case chat.ErrEmptyParticipants:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	resp := CreateChatResponse{
		ChatID: chatID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
