package familytreerepo

import (
	"errors"
	"family-tree/internal/core/familytree"

	"github.com/google/uuid"
	"github.com/mindstand/gogm/v2"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	ErrInvalidSessionMode  = errors.New("invalid session mode")
	ErrInvalidSessionValue = errors.New("invalid session value")
	ErrInvalidRelation     = errors.New("invalid relation type")
	ErrInvalidQueryResult  = errors.New("invalid query result")
	NodeConstraintMessage  = "because it still has relationships. To delete this node, you must first delete its relationships"
)

type Person struct {
	gogm.BaseUUIDNode

	Name     string    `gogm:"name=name" json:"-"`
	Parents  []*Person `gogm:"direction=incoming;relationship=PARENT" json:"-"`
	Children []*Person `gogm:"direction=outgoing;relationship=PARENT"`
	Spouse   *Person   `gogm:"direction=both;relationship=SPOUSE"`
}

func SessionMapper(sessionMode familytree.SessionMode) (neo4j.AccessMode, error) {
	switch sessionMode {
	case familytree.SessionRead:
		return neo4j.AccessModeRead, nil
	case familytree.SessionWrite:
		return neo4j.AccessModeWrite, nil
	default:
		return neo4j.AccessMode(0), ErrInvalidSessionMode
	}
}

func PersonMapper(person *Person) (*familytree.Person, error) {
	personUUID, err := uuid.Parse(person.UUID)
	if err != nil {
		return nil, err
	}
	newPerson := &familytree.Person{
		ID:   personUUID,
		Name: person.Name,
	}
	return newPerson, nil
}

func PeopleMapper(people ...*Person) ([]*familytree.Person, error) {
	mappedPeople := make([]*familytree.Person, 0, len(people))
	for _, person := range people {
		mappedPerson, err := PersonMapper(person)
		if err != nil {
			return nil, err
		}
		mappedPeople = append(mappedPeople, mappedPerson)
	}

	return mappedPeople, nil
}

type FamilyTree struct {
	People    map[string]familytree.Person
	Relations map[string][]FamilyTreeRelation
}
type FamilyTreeRelation struct {
	IncomingID   string
	RelationType familytree.RelationType
}

func FamilyTreeMapper(rootPerson Person) (*familytree.FamilyTree, error) {
	familyTree := &FamilyTree{
		People:    make(map[string]familytree.Person),
		Relations: make(map[string][]FamilyTreeRelation),
	}

	_, err := familyTreeMapperTraversal(rootPerson, familyTree)
	if err != nil {
		return nil, err
	}
	return reduceTree(familyTree)
}

func reduceTree(tree *FamilyTree) (*familytree.FamilyTree, error) {
	familyTree := &familytree.FamilyTree{
		People: []familytree.FamilyTreeNode{},
	}
	for _, person := range tree.People {
		newNode := familytree.FamilyTreeNode{
			Person: person,
		}
		relations := tree.Relations[person.ID.String()]
		if len(relations) > 0 {
			newNode.Relations = make([]familytree.FamilyTreeRelation, 0, len(relations))
		}
		for _, relation := range relations {
			incomingUUID, err := uuid.Parse(relation.IncomingID)
			if err != nil {
				return nil, err
			}
			newNode.Relations = append(newNode.Relations, familytree.FamilyTreeRelation{
				PersonID:     incomingUUID,
				RelationType: relation.RelationType,
			})
		}
		familyTree.People = append(familyTree.People, newNode)
	}

	return familyTree, nil
}

func familyTreeMapperTraversal(rootPerson Person, familyTree *FamilyTree) (familytree.Person, error) {
	person, err := PersonMapper(&rootPerson)
	if err != nil {
		return familytree.Person{}, err
	}
	familyTree.People[person.ID.String()] = *person
	//familyTree.Relations[person.ID.String()] = []familytree.FamilyTreeRelation{}
	if rootPerson.Spouse != nil {
		spouse, ok := familyTree.People[rootPerson.Spouse.UUID]
		if !ok {
			spouse, err = familyTreeMapperTraversal(*rootPerson.Spouse, familyTree)
			if err != nil {
				return familytree.Person{}, err
			}
			familyTree.Relations[person.ID.String()] = append(familyTree.Relations[person.ID.String()], FamilyTreeRelation{
				IncomingID:   spouse.ID.String(),
				RelationType: familytree.RelationTypeSpouse,
			})
		}
	}
	for _, child := range rootPerson.Children {
		mappedChild, ok := familyTree.People[child.UUID]
		if !ok {
			mappedChild, err = familyTreeMapperTraversal(*child, familyTree)
			if err != nil {
				return familytree.Person{}, err
			}
		}
		familyTree.Relations[person.ID.String()] = append(familyTree.Relations[person.ID.String()], FamilyTreeRelation{
			IncomingID:   mappedChild.ID.String(),
			RelationType: familytree.RelationTypeParent,
		})
	}
	for _, parent := range rootPerson.Parents {
		_, ok := familyTree.People[parent.UUID]
		if !ok {
			_, err := familyTreeMapperTraversal(*parent, familyTree)
			if err != nil {
				return familytree.Person{}, err
			}
		}
	}
	return *person, nil
}
