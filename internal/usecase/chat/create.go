package chat

import (
	"context"
	"database/sql"
	"errors"
	"sbChat/internal/repository"
)

var (
	ErrChatAlreadyExists = errors.New("chat already exists")
	ErrEmptyParticipants = errors.New("empty participants")
)

type CreateChatUseCase struct {
	chatRepo repository.ChatRepository
}

func NewCreateChatUseCase(chatRepo repository.ChatRepository) *CreateChatUseCase {
	return &CreateChatUseCase{chatRepo: chatRepo}
}

func (uc *CreateChatUseCase) Execute(ctx context.Context, participantIDs ...string) (string, error) {
	if len(participantIDs) == 0 {
		return "", ErrEmptyParticipants
	}
	
	// Проверяем, есть ли уже чат с такими участниками
	existingChatID, err := uc.chatRepo.FindChatByParticipants(ctx, participantIDs...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	if existingChatID != "" {
		return "", ErrChatAlreadyExists
	}

	// Создаем чат с участниками
	chatID, err := uc.chatRepo.CreateChatWithParticipants(ctx, participantIDs...)
	if err != nil {
		return "", err
	}

	return chatID, nil
}
