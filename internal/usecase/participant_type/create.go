package participant_type

import (
	"context"
	"errors"
	"fmt"
	"sbChat/internal/repository"
)

var (
	ErrEmptyTypeName      = errors.New("participant type name cannot be empty")
	ErrTypeCreationFailed = errors.New("failed to create participant type in storage")
	ErrTypeLookupFailed   = errors.New("failed to retrieve participant type information")
)

type CreateParticipantType struct {
	repo repository.ParticipantTypeRepository
}

func NewCreateParticipantTypeUseCase(repo repository.ParticipantTypeRepository) *CreateParticipantType {
	return &CreateParticipantType{repo: repo}
}

// Execute выполняет бизнес-логику для создания нового типа участника
// Возвращает ID созданного типа или ошибку, если операция не удалась
func (uc *CreateParticipantType) Execute(ctx context.Context, participantTypeName string) (string, error) {
	if participantTypeName == "" {
		return "", fmt.Errorf("%w: empty input received", ErrEmptyTypeName)
	}

	// Вызов репозитория для создания типа участника
	typeID, err := uc.repo.CreateParticipantType(ctx, participantTypeName)
	if err != nil {
		return "", fmt.Errorf("%w: %v | type: '%s'", ErrTypeCreationFailed, err, participantTypeName)
	}

	return typeID, nil
}
