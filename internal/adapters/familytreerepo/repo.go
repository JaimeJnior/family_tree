package familytreerepo

import (
	"context"
	"errors"
	"family-tree/internal/core/familytree"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mindstand/gogm/v2"
)

func NewFamilyTreeRepo(gogm *gogm.Gogm) *FamilyTreeRepo {
	return &FamilyTreeRepo{
		Gogm: gogm,
	}
}

type FamilyTreeRepo struct {
	Gogm *gogm.Gogm
}

func (repo *FamilyTreeRepo) OpenSession(ctx context.Context, mode familytree.SessionMode) (interface{}, error) {
	neo4jMode, err := SessionMapper(mode)
	if err != nil {
		return nil, err
	}
	session, err := repo.Gogm.NewSessionV2(gogm.SessionConfig{AccessMode: neo4jMode})
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (repo *FamilyTreeRepo) getSessionFromContext(ctx context.Context) (gogm.SessionV2, error) {
	genSession := ctx.Value(familytree.SessionKey)
	session, ok := genSession.(gogm.SessionV2)
	if !ok {
		return nil, ErrInvalidSessionValue
	}
	return session, nil
}

func (repo *FamilyTreeRepo) CloseSession(ctx context.Context) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return
	}

	session.Close()
}

func (repo *FamilyTreeRepo) getGogmPerson(ctx context.Context, id uuid.UUID) (*Person, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	person := &Person{}

	err = session.Load(ctx, person, id.String())
	if err != nil {
		if errors.Is(err, gogm.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return person, nil
}

func (repo *FamilyTreeRepo) GetPerson(ctx context.Context, id uuid.UUID) (*familytree.Person, error) {
	person, err := repo.getGogmPerson(ctx, id)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, nil
	}
	parsedUUID, err := uuid.Parse(person.BaseUUIDNode.UUID)
	if err != nil {
		return nil, err
	}
	return &familytree.Person{
		ID:   parsedUUID,
		Name: person.Name,
	}, err

}

func (repo *FamilyTreeRepo) SavePerson(ctx context.Context, person *familytree.Person) error {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return err
	}
	newPerson := &Person{
		Name: person.Name,
	}

	err = session.Save(ctx, newPerson)
	if err != nil {
		return err
	}
	person.ID, err = uuid.Parse(newPerson.BaseUUIDNode.UUID)
	return err
}

func (repo *FamilyTreeRepo) GetParents(ctx context.Context, personID uuid.UUID) ([]*familytree.Person, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		MATCH p=(person:Person {uuid:$uuid})<-[:PARENT]-(parent) 
		return p

	`
	person := &Person{}
	err = session.Query(context.Background(), query, map[string]interface{}{"uuid": personID.String()}, person)
	if err != nil {
		if errors.Is(err, gogm.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return PeopleMapper(person.Parents...)
}

func (repo *FamilyTreeRepo) SaveRelation(ctx context.Context, relation familytree.PersonRelation) error {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		MATCH (a:Person),
			  (b:Person)
		WHERE a.uuid = $top AND b.uuid = $bottom
		CREATE (a)-[:%s]->(b)
	`, relation.RelationType)
	_, _, err = session.QueryRaw(ctx, query, map[string]interface{}{
		"top":    relation.Top.ID.String(),
		"bottom": relation.Bottom.ID.String(),
	})

	return err

}

func (repo *FamilyTreeRepo) GetLowestCommonAncestor(ctx context.Context, firstPerson familytree.Person, secondPerson familytree.Person) (*familytree.Person, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		MATCH (c1:Person {uuid: $firstPerson})<-[:PARENT*0..]-(p1:Person)
		MATCH (:Person {uuid: $secondPerson})<-[:PARENT*0..]-(p2:Person)
		WHERE p1.uuid = p2.uuid
		MATCH path = (c1)<-[:PARENT*0..]-(p1)
		RETURN p1
		ORDER BY length(path)
		LIMIT 1
	`
	ancestor := &Person{}
	err = session.Query(ctx, query, map[string]interface{}{
		"firstPerson":  firstPerson.ID.String(),
		"secondPerson": secondPerson.ID.String(),
	}, ancestor)
	if err != nil {
		if errors.Is(err, gogm.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return PersonMapper(ancestor)
}

func (repo *FamilyTreeRepo) GetPeople(ctx context.Context, pagination familytree.PaginationDetails) (*familytree.PeopleList, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	countQuery := `
		MATCH (person:Person) 
		RETURN count(person)
	`
	totalItens := int64(0)
	result, _, err := session.QueryRaw(ctx, countQuery, map[string]interface{}{})

	if err != nil {
		return nil, err
	}
	totalItens, ok := result[0][0].(int64)
	if !ok {
		return nil, ErrInvalidQueryResult
	}
	peopleList := &familytree.PeopleList{
		Metadata: familytree.ListMetadata{
			TotalItens: int(totalItens),
			Page:       pagination.Page,
		},
	}
	query := `
		MATCH (person:Person) 
		RETURN person
		SKIP $skip
		LIMIT $pagesize
	`
	var rawList []*Person
	session.Query(ctx, query, map[string]interface{}{
		"skip":     pagination.Page * pagination.PageSize,
		"pagesize": pagination.PageSize,
	}, &rawList)
	if err != nil {
		return nil, err
	}
	mappedList, err := PeopleMapper(rawList...)
	if err != nil {
		return nil, err
	}
	peopleList.Content = mappedList
	return peopleList, nil
}

func (repo *FamilyTreeRepo) GetFamilyTree(ctx context.Context, person familytree.Person) (*familytree.FamilyTree, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rootPerson := &Person{}
	query := `
		MATCH (startingNode:Person {uuid:$uuid})<-[t:PARENT*0..]-(parent)
		OPTIONAL MATCH (parent)-[s:SPOUSE]-(spouse)-[:PARENT*..]->(startingNode)
		return parent as person, t as relation, s as spouse
		UNION
		MATCH (child)<-[t:PARENT*0..]-(startingNode:Person {uuid:$uuid})
		RETURN child as person, t as relation, null as spouse
		UNION
		MATCH (startingNode:Person {uuid:$uuid})<-[:PARENT]-()-[t:PARENT]->(sibling)
		RETURN sibling as person, t as relation, null as spouse
		UNION
		MATCH (startingNode:Person {uuid:$uuid})<-[:PARENT]-()-[:PARENT]->()-[t:PARENT]->(nephew)
		RETURN nephew as person, t as relation, null as spouse
	`
	err = session.Query(ctx, query, map[string]interface{}{
		"uuid": person.ID.String(),
	}, rootPerson)
	if err != nil {
		return nil, err
	}
	return FamilyTreeMapper(*rootPerson)
}

func (repo *FamilyTreeRepo) GetShortestPathLength(ctx context.Context, firstPerson familytree.Person, secondPerson familytree.Person) (int, bool, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return 0, false, err
	}
	queryRaw := `
	MATCH
		(first:Person {uuid: $uuid_first}),
		(second:Person {uuid: $uuid_second}),
		p = shortestPath((first)-[*..]-(second))
	RETURN length(p)
	`
	result, _, err := session.QueryRaw(ctx, queryRaw, map[string]interface{}{
		"uuid_first":  firstPerson.ID.String(),
		"uuid_second": secondPerson.ID.String(),
	})

	if err != nil {
		return 0, false, err
	}
	if len(result) == 0 {
		return 0, false, nil
	}
	totalItens, ok := result[0][0].(int64)
	if !ok {
		return 0, false, ErrInvalidQueryResult
	}
	return int(totalItens), true, nil
}

func (repo *FamilyTreeRepo) HasCommonChild(ctx context.Context, firstPerson familytree.Person, secondPerson familytree.Person) (bool, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	queryRaw := `
	MATCH
		(first:Person {uuid: $uuid_first}),
		(second:Person {uuid: $uuid_second})
		RETURN exists( (first)-[:PARENT]->()<-[:PARENT]-(second) )
	`
	result, _, err := session.QueryRaw(ctx, queryRaw, map[string]interface{}{
		"uuid_first":  firstPerson.ID.String(),
		"uuid_second": secondPerson.ID.String(),
	})

	if err != nil {
		return false, err
	}
	hasChild, ok := result[0][0].(bool)
	if !ok {
		return false, ErrInvalidQueryResult
	}
	return hasChild, nil
}

func (repo *FamilyTreeRepo) GetSpouse(ctx context.Context, person familytree.Person) (*familytree.Person, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	spouse := &Person{}
	query := `
	MATCH
	(person:Person {uuid: $uuid}),
	(person)-[:SPOUSE]-(spouse)
	RETURN spouse
	`
	err = session.Query(ctx, query, map[string]interface{}{
		"uuid": person.ID.String(),
	}, spouse)
	if err != nil {
		if errors.Is(err, gogm.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	mappedSpouse, err := PersonMapper(spouse)
	if err != nil {
		return nil, err
	}
	return mappedSpouse, nil
}

func (repo *FamilyTreeRepo) GetParentMaritalChildCount(ctx context.Context, person familytree.Person) (int, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	queryRaw := `
	MATCH (father)-[:PARENT]->(:Person {uuid : $uuid})<-[:PARENT]-(mother)
	MATCH (father)-[:SPOUSE]->(mother)
	MATCH (father)-[:PARENT]->(sibling:Person)<-[:PARENT]-(mother)
	RETURN count(sibling)
	`
	result, _, err := session.QueryRaw(ctx, queryRaw, map[string]interface{}{
		"uuid": person.ID.String(),
	})

	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	totalItens, ok := result[0][0].(int64)
	if !ok {
		return 0, ErrInvalidSessionValue
	}
	return int(totalItens), nil
}
func (repo *FamilyTreeRepo) formatRelation(relationType familytree.RelationType) string {
	relationString := "-[r:%s]-"
	if relationType.Directional {
		relationString = "-[r:%s]->"
	}
	return fmt.Sprintf(relationString, relationType)
}

func (repo *FamilyTreeRepo) DeleteRelationship(ctx context.Context, firstPerson familytree.Person, secondPerson familytree.Person, relationType familytree.RelationType) (bool, error) {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	queryRaw := fmt.Sprintf(`
	MATCH (:Person {uuid : $first_uuid})%s(:Person {uuid : $second_uuid})
	DELETE r
	RETURN count(r)
	`, repo.formatRelation(relationType))
	result, _, err := session.QueryRaw(ctx, queryRaw, map[string]interface{}{
		"first_uuid":  firstPerson.ID.String(),
		"second_uuid": secondPerson.ID.String(),
	})
	if err != nil {
		return false, err
	}
	if len(result) == 0 {
		return false, nil
	}
	deletedItens, ok := result[0][0].(int64)
	if !ok {
		return false, ErrInvalidSessionValue
	}
	return deletedItens > 0, nil
}
func (repo *FamilyTreeRepo) DeletePerson(ctx context.Context, person familytree.Person) error {
	session, err := repo.getSessionFromContext(ctx)
	if err != nil {
		return err
	}
	queryRaw := `
	MATCH (person:Person {uuid : $uuid})
	DELETE person
	`
	_, _, err = session.QueryRaw(ctx, queryRaw, map[string]interface{}{
		"uuid": person.ID.String(),
	})
	if err != nil {
		if strings.Contains(err.Error(), NodeConstraintMessage) {
			return familytree.ErrPersonStillHasRelations
		}
		return err
	}
	return nil
}
