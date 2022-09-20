package server

import (
	"encoding/xml"
	"errors"
	"family-tree/internal/core/familytree"
	"net/http"

	"github.com/google/uuid"
)

const (
	PaginationPageParam = "page"
	PaginationSizeParam = "size"
)

var (
	ErrNotUUID              = errors.New("invalid uuid")
	ErrNoPathFound          = errors.New("no path found between people")
	AcceptApplicationJson   = "application/json"
	AcceptApplicationXML    = "application/xml"
	AcceptApplicationBinary = "binary"
	AcceptOctetStream       = "application/octet-stream"
	ErrorStatusResponseMap  = map[error]int{
		familytree.ErrCreateNilPerson:           http.StatusBadRequest,
		familytree.ErrEmptyPersonName:           http.StatusBadRequest,
		familytree.ErrDuplicateRelation:         http.StatusBadRequest,
		familytree.ErrMaxParents:                http.StatusBadRequest,
		familytree.ErrSameParentChildID:         http.StatusBadRequest,
		familytree.ErrIncestuousRelation:        http.StatusBadRequest,
		familytree.ErrPersonNotFound:            http.StatusNotFound,
		familytree.ErrCoupleHasNoChild:          http.StatusBadRequest,
		familytree.ErrHasSpouseAlready:          http.StatusBadRequest,
		familytree.ErrRelationNotFound:          http.StatusNotFound,
		familytree.ErrOnlyChildFromSpouseCouple: http.StatusBadRequest,
		familytree.ErrPersonStillHasRelations:   http.StatusBadRequest,
	}
)

type Person struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Error struct {
	Message string `json:"message"`
}

type PostPersonRequest struct {
	Name string `json:"name"`
}
type PostCreateParentRelationshipRequest struct {
	ParentID uuid.UUID `json:"parentID"`
	ChildID  uuid.UUID `json:"childID"`
}

func (r PostCreateParentRelationshipRequest) Validate() error {
	if r.ParentID == uuid.Nil {
		return ErrNotUUID
	}
	if r.ChildID == uuid.Nil {
		return ErrNotUUID
	}
	return nil
}

type DeleteParentRelationshipRequest struct {
	ParentID uuid.UUID `json:"parentID"`
	ChildID  uuid.UUID `json:"childID"`
}

func (r DeleteParentRelationshipRequest) Validate() error {
	if r.ParentID == uuid.Nil {
		return ErrNotUUID
	}
	if r.ChildID == uuid.Nil {
		return ErrNotUUID
	}
	return nil
}

type PostCreateSpouseRelationshipRequest struct {
	FirstSpouseID  uuid.UUID `json:"firstSpouseID"`
	SecondSpouseID uuid.UUID `json:"secondSpouseID"`
}

func (r PostCreateSpouseRelationshipRequest) Validate() error {
	if r.FirstSpouseID == uuid.Nil {
		return ErrNotUUID
	}
	if r.SecondSpouseID == uuid.Nil {
		return ErrNotUUID
	}
	return nil
}

type DeleteSpouseRelationshipRequest struct {
	FirstSpouseID  uuid.UUID `json:"firstSpouseID"`
	SecondSpouseID uuid.UUID `json:"secondSpouseID"`
}

func (r DeleteSpouseRelationshipRequest) Validate() error {
	if r.FirstSpouseID == uuid.Nil {
		return ErrNotUUID
	}
	if r.SecondSpouseID == uuid.Nil {
		return ErrNotUUID
	}
	return nil
}

type GetBaconsNumberResponse struct {
	PathLength int `json:"pathLength"`
}

type PaginationResponseMetadata struct {
	Page       int `json:"page"`
	TotalItens int `json:"totalItens"`
}
type GetPeopleResponse struct {
	Content  []*Person                  `json:"content"`
	Metadata PaginationResponseMetadata `json:"metadata"`
}

type FamilyTreeRelation struct {
	PersonID     uuid.UUID `json:"relativeID" xml:"id,attr"`
	RelationType string    `json:"relation" xml:"relationType"`
}

type FamilyTreeNode struct {
	Person    Person               `json:"personID" xml:"person"`
	Relations []FamilyTreeRelation `json:"relations,omitempty" xml:"relations"`
}

type FamilyTree struct {
	XMLName xml.Name         `json:"-" xml:"familyTree"`
	People  []FamilyTreeNode `json:"people" xml:"people"`
}

func FamilyTreeMapper(tree *familytree.FamilyTree) *FamilyTree {
	if tree == nil {
		return nil
	}
	newTree := &FamilyTree{
		People: make([]FamilyTreeNode, 0, len(tree.People)),
	}
	for _, node := range tree.People {
		convertedNode := &FamilyTreeNode{
			Person:    Person(node.Person),
			Relations: make([]FamilyTreeRelation, 0, len(node.Relations)),
		}
		for _, relation := range node.Relations {
			convertedNode.Relations = append(convertedNode.Relations, FamilyTreeRelation{
				PersonID:     relation.PersonID,
				RelationType: relation.RelationType.String(),
			})
		}
		newTree.People = append(newTree.People, *convertedNode)

	}
	return newTree
}

func PersonMapper(person familytree.Person) Person {
	return Person(person)
}

func PeopleMapper(people []*familytree.Person) []*Person {
	mappedPeople := make([]*Person, 0, len(people))
	for _, person := range people {
		mappedPerson := Person(*person)
		mappedPeople = append(mappedPeople, &mappedPerson)

	}
	return mappedPeople
}
