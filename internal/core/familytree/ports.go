package familytree

import (
	"context"

	"github.com/google/uuid"
)

type FamilyTreeRepo interface {
	OpenSession(ctx context.Context, mode SessionMode) (interface{}, error)
	CloseSession(ctx context.Context)
	SavePerson(ctx context.Context, person *Person) error
	GetPerson(ctx context.Context, personID uuid.UUID) (*Person, error)
	SaveRelation(ctx context.Context, relation PersonRelation) error
	GetParents(ctx context.Context, personID uuid.UUID) ([]*Person, error)
	GetLowestCommonAncestor(ctx context.Context, firstPerson Person, secondPerson Person) (*Person, error)
	GetPeople(ctx context.Context, pagination PaginationDetails) (*PeopleList, error)
	GetFamilyTree(ctx context.Context, person Person) (*FamilyTree, error)
	GetShortestPathLength(ctx context.Context, firstPerson Person, secondPerson Person) (int, bool, error)
	HasCommonChild(ctx context.Context, firstPerson Person, secondPerson Person) (bool, error)
	GetSpouse(ctx context.Context, person Person) (*Person, error)
	GetParentMaritalChildCount(ctx context.Context, person Person) (int, error)
	DeleteRelationship(ctx context.Context, firstPerson Person, secondPerson Person, relationType RelationType) (bool, error)
	DeletePerson(ctx context.Context, person Person) error
}

type PersonUseCasePort interface {
	CreatePerson(ctx context.Context, person *Person) error
	GetPeople(ctx context.Context, pagination PaginationDetails) (*PeopleList, error)
	GetPerson(ctx context.Context, personID uuid.UUID) (*Person, error)
	GetBaconsNumber(ctx context.Context, firstPersonID uuid.UUID, secondPersonID uuid.UUID) (int, bool, error)
	DeletePerson(ctx context.Context, personID uuid.UUID) error
}

type RelationshipUseCasePort interface {
	CreateParentRelation(ctx context.Context, parentID uuid.UUID, childID uuid.UUID) error
	CreateSpouseRelation(ctx context.Context, firstSpouseID uuid.UUID, secondSpouseID uuid.UUID) error
	GetFamilyTree(ctx context.Context, personID uuid.UUID) (*FamilyTree, error)
	DeleteSpouseRelation(ctx context.Context, firstSpouseID uuid.UUID, secondSpouseID uuid.UUID) error
	DeleteParentRelation(ctx context.Context, parentID uuid.UUID, childID uuid.UUID) error
}
