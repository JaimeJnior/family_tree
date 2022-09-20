package familytree

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type PersonUseCase struct {
	familyTreeRepo FamilyTreeRepo
}

func NewPersonUseCase(familyTreeRepo FamilyTreeRepo) *PersonUseCase {

	return &PersonUseCase{
		familyTreeRepo: familyTreeRepo,
	}
}

func (useCase *PersonUseCase) openSession(ctx context.Context, sessionMode SessionMode) (context.Context, error) {
	session, err := useCase.familyTreeRepo.OpenSession(ctx, sessionMode)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, SessionKey, session), nil
}

func (useCase *PersonUseCase) CreatePerson(ctx context.Context, person *Person) error {
	newCtx, err := useCase.openSession(ctx, SessionWrite)
	if err != nil {
		return err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)
	if person == nil {
		return ErrCreateNilPerson
	}
	trimmedName := strings.TrimSpace(person.Name)
	if trimmedName == "" {
		return ErrEmptyPersonName
	}
	person.Name = trimmedName
	return useCase.familyTreeRepo.SavePerson(ctx, person)
}

func (useCase *PersonUseCase) paginationValidate(pagination *PaginationDetails) {
	if pagination.Page < 0 {
		pagination.Page = GetPeopleDefaultPage
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = GetPeopleDefaultPageSize
	}
	if pagination.PageSize > GetPeopleMaxPageSize {
		pagination.PageSize = GetPeopleMaxPageSize
	}
}

func (useCase *PersonUseCase) GetPeople(ctx context.Context, pagination PaginationDetails) (*PeopleList, error) {
	newCtx, err := useCase.openSession(ctx, SessionRead)
	if err != nil {
		return nil, err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)
	useCase.paginationValidate(&pagination)
	return useCase.familyTreeRepo.GetPeople(ctx, pagination)

}

func (useCase *PersonUseCase) GetPerson(ctx context.Context, personID uuid.UUID) (*Person, error) {
	newCtx, err := useCase.openSession(ctx, SessionRead)
	if err != nil {
		return nil, err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)
	return useCase.familyTreeRepo.GetPerson(ctx, personID)
}

func (useCase *PersonUseCase) GetBaconsNumber(ctx context.Context, firstPersonID uuid.UUID, secondPersonID uuid.UUID) (int, bool, error) {
	newCtx, err := useCase.openSession(ctx, SessionRead)
	if err != nil {
		return 0, false, err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)

	firstPerson, err := useCase.familyTreeRepo.GetPerson(ctx, firstPersonID)
	if err != nil {
		return 0, false, err
	}
	if firstPerson == nil {
		return 0, false, ErrPersonNotFound
	}

	secondPerson, err := useCase.familyTreeRepo.GetPerson(ctx, secondPersonID)
	if err != nil {
		return 0, false, err
	}
	if secondPerson == nil {
		return 0, false, ErrPersonNotFound
	}
	if firstPerson.ID == secondPerson.ID {
		return 0, true, nil
	}

	return useCase.familyTreeRepo.GetShortestPathLength(ctx, *firstPerson, *secondPerson)
}

func (useCase *PersonUseCase) DeletePerson(ctx context.Context, personID uuid.UUID) error {
	newCtx, err := useCase.openSession(ctx, SessionWrite)
	if err != nil {
		return err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)

	person, err := useCase.familyTreeRepo.GetPerson(ctx, personID)
	if err != nil {
		return err
	}
	if person == nil {
		return ErrPersonNotFound
	}
	return useCase.familyTreeRepo.DeletePerson(ctx, *person)

}
