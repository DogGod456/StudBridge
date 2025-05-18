package chat

import (
	"context"
	"errors"
	"fmt"
	"sbChat/internal/repository"
	"sbChat/internal/usecase/participant"
)

var (
	ErrEmptyParticipants      = errors.New("chat creation requires at least one participant")
	ErrParticipantProcessing  = errors.New("failed to process participant configuration")
	ErrChatCreationFailed     = errors.New("chat creation operation could not be completed")
	ErrInvalidParticipantData = errors.New("participant data validation failed")
)

// ParticipantInput определяет структуру участника чата
type ParticipantInput struct {
	ID   string `json:"id"` // ref_id
	Type string `json:"type"`
}

type CreateChatUseCase struct {
	chatRepo           repository.ChatRepository
	participantCreator *participant.ParticipantUseCase
}

func NewCreateChatUseCase(
	chatRepo repository.ChatRepository,
	participantCreator *participant.ParticipantUseCase,
) *CreateChatUseCase {
	return &CreateChatUseCase{
		chatRepo:           chatRepo,
		participantCreator: participantCreator,
	}
}

func (uc *CreateChatUseCase) Execute(ctx context.Context, participants []ParticipantInput) (string, error) {
	if len(participants) == 0 {
		return "", fmt.Errorf("%w: zero participants provided", ErrEmptyParticipants)
	}

	participantIDs := make([]string, 0, len(participants))
	for i, p := range participants {
		if p.ID == "" || p.Type == "" {
			return "", fmt.Errorf("%w: participant %d has empty ID or Type", ErrInvalidParticipantData, i+1)
		}

		pid, err := uc.participantCreator.CreateOrGetParticipant(ctx, p.Type, p.ID) // p.Type и p.ID берутся из структуры ParticipantInput
		if err != nil {
			return "", fmt.Errorf("%w: [participant %d: %s/%s]: %v", ErrParticipantProcessing, i+1, p.Type, p.ID, err)
		}
		participantIDs = append(participantIDs, pid)
	}

	// Создаем чат с участниками
	chatID, err := uc.chatRepo.CreateChatWithParticipants(ctx, participantIDs...)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrChatCreationFailed, err)
	}

	return chatID, nil
}
