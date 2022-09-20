package familytree

import (
	"context"

	"github.com/google/uuid"
)

type RelationshipUseCase struct {
	familyTreeRepo FamilyTreeRepo
}

func NewRelationshipUseCase(familyTreeRepo FamilyTreeRepo) *RelationshipUseCase {

	return &RelationshipUseCase{
		familyTreeRepo: familyTreeRepo,
	}
}

func (useCase *RelationshipUseCase) openSession(ctx context.Context, sessionMode SessionMode) (context.Context, error) {
	session, err := useCase.familyTreeRepo.OpenSession(ctx, sessionMode)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, SessionKey, session), nil
}

func (useCase *RelationshipUseCase) validateCreateChildRelation(ctx context.Context, parent *Person, child *Person) error {
	parents, err := useCase.familyTreeRepo.GetParents(ctx, child.ID)
	if err != nil {
		return err
	}
	if len(parents) == MaxParents {
		return ErrMaxParents
	}
	if parent.ID == child.ID {
		return ErrSameParentChildID
	}
	for _, currentParent := range parents {
		if currentParent.ID == parent.ID {
			return ErrDuplicateRelation
		}
	}

	//Check incestuous relationship
	ancestor, err := useCase.familyTreeRepo.GetLowestCommonAncestor(ctx, *parent, *child)
	if err != nil {
		return err
	}
	if ancestor != nil {
		return ErrIncestuousRelation
	}

	return nil
}

func (useCase *RelationshipUseCase) validateSpouseExists(ctx context.Context, foundSpouse *Person, newSpouse *Person) error {
	if foundSpouse == nil {
		return nil
	}
	if foundSpouse.ID == newSpouse.ID {
		return ErrDuplicateRelation
	}
	return ErrHasSpouseAlready
}

func (useCase *RelationshipUseCase) validateCreateSpouseRelation(ctx context.Context, firstSpouse *Person, secondSpouse *Person) error {
	hasChild, err := useCase.familyTreeRepo.HasCommonChild(ctx, *firstSpouse, *secondSpouse)
	if err != nil {
		return err
	}
	if !hasChild {
		return ErrCoupleHasNoChild
	}
	firstSpouseCheck, err := useCase.familyTreeRepo.GetSpouse(ctx, *firstSpouse)
	if err != nil {
		return err
	}
	err = useCase.validateSpouseExists(ctx, firstSpouseCheck, secondSpouse)
	if err != nil {
		return err
	}
	secondSpouseCheck, err := useCase.familyTreeRepo.GetSpouse(ctx, *secondSpouse)
	if err != nil {
		return err
	}
	return useCase.validateSpouseExists(ctx, secondSpouseCheck, firstSpouse)
}

func (useCase *RelationshipUseCase) CreateParentRelation(ctx context.Context, parentID uuid.UUID, childID uuid.UUID) error {

	newCtx, err := useCase.openSession(ctx, SessionWrite)
	if err != nil {
		return err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)
	parent, err := useCase.familyTreeRepo.GetPerson(ctx, parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return ErrPersonNotFound
	}
	child, err := useCase.familyTreeRepo.GetPerson(ctx, childID)
	if err != nil {
		return err
	}
	if child == nil {
		return ErrPersonNotFound
	}

	if err := useCase.validateCreateChildRelation(ctx, parent, child); err != nil {
		return err
	}

	err = useCase.familyTreeRepo.SaveRelation(ctx, PersonRelation{
		Top:          *parent,
		Bottom:       *child,
		RelationType: RelationTypeParent,
	})
	return err
}

func (useCase *RelationshipUseCase) CreateSpouseRelation(ctx context.Context, firstSpouseID uuid.UUID, secondSpouseID uuid.UUID) error {

	newCtx, err := useCase.openSession(ctx, SessionWrite)
	if err != nil {
		return err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)
	firstSpouse, err := useCase.familyTreeRepo.GetPerson(ctx, firstSpouseID)
	if err != nil {
		return err
	}
	if firstSpouse == nil {
		return ErrPersonNotFound
	}
	secondSpouse, err := useCase.familyTreeRepo.GetPerson(ctx, secondSpouseID)
	if err != nil {
		return err
	}
	if secondSpouse == nil {
		return ErrPersonNotFound
	}

	if err := useCase.validateCreateSpouseRelation(ctx, firstSpouse, secondSpouse); err != nil {
		return err
	}

	err = useCase.familyTreeRepo.SaveRelation(ctx, PersonRelation{
		Top:          *firstSpouse,
		Bottom:       *secondSpouse,
		RelationType: RelationTypeSpouse,
	})
	return err
}

func (useCase *RelationshipUseCase) GetFamilyTree(ctx context.Context, personID uuid.UUID) (*FamilyTree, error) {
	newCtx, err := useCase.openSession(ctx, SessionRead)
	if err != nil {
		return nil, err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)

	person, err := useCase.familyTreeRepo.GetPerson(ctx, personID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}

	return useCase.familyTreeRepo.GetFamilyTree(ctx, *person)
}

func (useCase *RelationshipUseCase) DeleteParentRelation(ctx context.Context, parentID uuid.UUID, childID uuid.UUID) error {
	newCtx, err := useCase.openSession(ctx, SessionWrite)
	if err != nil {
		return err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)

	parent, err := useCase.familyTreeRepo.GetPerson(ctx, parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return ErrPersonNotFound
	}
	child, err := useCase.familyTreeRepo.GetPerson(ctx, childID)
	if err != nil {
		return err
	}
	if child == nil {
		return ErrPersonNotFound
	}
	count, err := useCase.familyTreeRepo.GetParentMaritalChildCount(ctx, *child)
	if err != nil {
		return err
	}
	if count == 1 {
		return ErrOnlyChildFromSpouseCouple
	}
	ok, err := useCase.familyTreeRepo.DeleteRelationship(ctx, *parent, *child, RelationTypeParent)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRelationNotFound
	}
	return nil
}

func (useCase *RelationshipUseCase) DeleteSpouseRelation(ctx context.Context, firstSpouseID uuid.UUID, secondSpouseID uuid.UUID) error {
	newCtx, err := useCase.openSession(ctx, SessionWrite)
	if err != nil {
		return err
	}
	ctx = newCtx
	defer useCase.familyTreeRepo.CloseSession(ctx)

	firstSpouse, err := useCase.familyTreeRepo.GetPerson(ctx, firstSpouseID)
	if err != nil {
		return err
	}
	if firstSpouse == nil {
		return ErrPersonNotFound
	}
	secondSpouse, err := useCase.familyTreeRepo.GetPerson(ctx, secondSpouseID)
	if err != nil {
		return err
	}
	if secondSpouse == nil {
		return ErrPersonNotFound
	}
	ok, err := useCase.familyTreeRepo.DeleteRelationship(ctx, *firstSpouse, *secondSpouse, RelationTypeSpouse)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRelationNotFound
	}
	return nil
}
