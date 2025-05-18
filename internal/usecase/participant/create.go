package participant

import (
	"context"
	"errors"
	"fmt"
	"sbChat/internal/repository"
	participant_type "sbChat/internal/usecase/participant_type"
)

var (
	ErrInvalidInput      = errors.New("invalid input parameters")
	ErrParticipantExists = errors.New("participant already exists with different type")
	ErrTypeRetrieval     = errors.New("failed to retrieve participant type ID from repository")
	ErrTypeCreation      = errors.New("failed to create new participant type")
	ErrParticipantCheck  = errors.New("error checking existing participant by reference ID")
	ErrTypeVerification  = errors.New("error verifying participant type consistency")
	ErrParticipantCreate = errors.New("failed to create new participant in repository")
)

type ParticipantUseCase struct {
	repo                   repository.ParticipantRepository
	participantTypeCreator *participant_type.CreateParticipantType
}

func NewParticipantUseCase(
	repo repository.ParticipantRepository,
	participantTypeCreator *participant_type.CreateParticipantType,
) *ParticipantUseCase {
	return &ParticipantUseCase{
		repo:                   repo,
		participantTypeCreator: participantTypeCreator,
	}
}

func (uc *ParticipantUseCase) CreateOrGetParticipant(ctx context.Context, participantTypeName, refID string) (string, error) {
	// Валидация входных данных
	if participantTypeName == "" || refID == "" {
		return "", fmt.Errorf("%w: participant type name or reference ID cannot be empty", ErrInvalidInput)
	}

	// Получаем ID типа участника
	typeID, err := uc.repo.GetParticipantTypeIDByTypeName(ctx, participantTypeName)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTypeRetrieval, err)
	}

	// Если тип не существует, создаем его
	if typeID == "" {
		typeID, err = uc.participantTypeCreator.Execute(ctx, participantTypeName)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrTypeCreation, err)
		}

	}

	// Проверяем существование участника по refID
	participantID, err := uc.repo.GetParticipantIdByRef(ctx, refID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrParticipantCheck, err)
	}

	if participantID != "" {
		// Проверяем, совпадает ли тип участника
		exists, err := uc.repo.ExistsParticipantByIDAndType(ctx, participantID, typeID)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrTypeVerification, err)
		}
		if exists {
			return participantID, nil
		}
	}

	// Создаем нового участника
	participantID, err = uc.repo.CreateParticipant(ctx, typeID, refID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrParticipantCreate, err)
	}

	return participantID, nil
}
