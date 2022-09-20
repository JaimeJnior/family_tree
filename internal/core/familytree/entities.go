package familytree

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type RelationType struct {
	Name        string
	Directional bool
}

func (rel RelationType) String() string {
	return rel.Name
}

type SessionMode string
type ContextKey string

const (
	MaxParents               = 2
	SessionRead              = SessionMode("READ")
	SessionWrite             = SessionMode("WRITE")
	SessionKey               = ContextKey("family_tree_session")
	GetPeopleMaxPageSize     = 50
	GetPeopleDefaultPageSize = 10
	GetPeopleDefaultPage     = 0
)

var (
	RelationTypeParent           = RelationType{"PARENT", true}
	RelationTypeSpouse           = RelationType{"SPOUSE", false}
	ErrCreateNilPerson           = errors.New("can't create nil person")
	ErrEmptyPersonName           = errors.New("person name can't be empty")
	ErrDuplicateRelation         = errors.New("relation already exists")
	ErrMaxParents                = fmt.Errorf("child already has %d parents", MaxParents)
	ErrSameParentChildID         = errors.New("child and parent ids can't be the same")
	ErrIncestuousRelation        = errors.New("child and parent are already relatives")
	ErrPersonNotFound            = errors.New("person not found")
	ErrCoupleHasNoChild          = errors.New("couple has no common child")
	ErrHasSpouseAlready          = errors.New("person has spouse already")
	ErrOnlyChildFromSpouseCouple = errors.New("can't delete only child relation of spouse coupe")
	ErrRelationNotFound          = errors.New("relation not found")
	ErrPersonStillHasRelations   = errors.New("person still has relations")
)

type PaginationDetails struct {
	Page     int
	PageSize int
}

type ListMetadata struct {
	Page       int
	TotalItens int
}
type PeopleList struct {
	Content  []*Person
	Metadata ListMetadata
}

type Person struct {
	ID   uuid.UUID
	Name string
}

type PersonRelation struct {
	Top          Person
	Bottom       Person
	RelationType RelationType
}

type FamilyTreeRelation struct {
	PersonID     uuid.UUID
	RelationType RelationType
}
type FamilyTreeNode struct {
	Person    Person
	Relations []FamilyTreeRelation
}

type FamilyTree struct {
	People []FamilyTreeNode
}
