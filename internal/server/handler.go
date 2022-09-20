package server

import (
	"encoding/json"
	"family-tree/internal/core/familytree"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetBaconsNumber godoc
// @Summary Busca o número de Bacon entre duas pessoas
// @Description Busca todas as pessoas salvas no banco
// @Tags person
// @Produce  json
// @Param personID path string true "ID da pessoa no formato uuid (XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX)"
// @Param targetID path string true "ID da pessoa alvo no formato uuid (XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX)"
// @Success 200 {object} GetBaconsNumberResponse
// @Router /person/{personID}/bacons/{targetID} [get]
func (server *Server) GetBaconsNumber(w http.ResponseWriter, r *http.Request) {
	stringUUID := chi.URLParam(r, "personID")
	personID, err := uuid.Parse(stringUUID)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, ErrNotUUID)
		return
	}
	stringUUID = chi.URLParam(r, "targetPersonID")
	targetPersonID, err := uuid.Parse(stringUUID)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, ErrNotUUID)
		return
	}

	baconsNumber, ok, err := server.PersonUseCase.GetBaconsNumber(r.Context(), personID, targetPersonID)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	if !ok {
		WriteErrorMessage(w, r, http.StatusNotFound, ErrNoPathFound)
		return
	}

	WriteJsonBody(w, r, http.StatusOK, GetBaconsNumberResponse{PathLength: baconsNumber})
}

// GetPerson godoc
// @Summary Busca detalhes de uma pessoa pelo seu id
// @Description Busca detalhes de uma pessoa pelo seu id
// @Description Retorna 404 caso não existe
// @Tags person
// @Produce  json
// @Param personID path string true "ID da pessoa no formato uuid (XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX)"
// @Success 200 {object} Person
// @Router /person/{personID} [get]
func (server *Server) GetPersonHandler(w http.ResponseWriter, r *http.Request) {
	stringUUID := chi.URLParam(r, "personID")
	personID, err := uuid.Parse(stringUUID)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, ErrNotUUID)
		return
	}

	person, err := server.PersonUseCase.GetPerson(r.Context(), personID)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, ErrNotUUID)
		return
	}
	if person == nil {
		WriteError(w, r, http.StatusNotFound)
		return
	}
	WriteJsonBody(w, r, http.StatusOK, PersonMapper(*person))
}

func (server *Server) DeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	stringUUID := chi.URLParam(r, "personID")
	personID, err := uuid.Parse(stringUUID)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, ErrNotUUID)
		return
	}

	err = server.PersonUseCase.DeletePerson(r.Context(), personID)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetFamilyTree godoc
// @Summary Busca a árvore genealógica de uma pessoa
// @Description Busca a árvore genealógica de uma pessoa, reduzindo relações redundantes
// @Description Resultado pode ser entregue tanto de json, xml e em binário
// @Description A relação de PARENT indica que a pessoa é pai da pessoa indicada
// @Description A relação de SPOUSE indica que a pesoa possui uma relação de casamento com a pessoa indica
// @Description Lembrando que para reduzir redundância a relação só aparece em uma das pessoas
// @Description Na árvore está incluso:
// @Description a) Todos os seus ancestrais
// @Description b) Seus filhos
// @Description c) Seus sobrinhos
// @Description d) Relações de esposo entre ancestrais, não incluindo se estiver a relação de esposo com alguém sem relação sanguínea
// @Tags relationship
// @Produce  json
// @Produce  application/xml
// @Produce  octet-stream
// @Param personID path string true "ID da pessoa no formato uuid (XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX)"
// @Success 200 {object} FamilyTree
// @Router /person/{personID}/tree [get]
func (server *Server) GetFamilyTree(w http.ResponseWriter, r *http.Request) {
	stringUUID := chi.URLParam(r, "personID")
	personID, err := uuid.Parse(stringUUID)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, ErrNotUUID)
		return
	}

	familyTree, err := server.RelationshipUseCase.GetFamilyTree(r.Context(), personID)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	if familyTree == nil {
		WriteError(w, r, http.StatusNotFound)
		return
	}
	ResponseStrategy(r.Header.Values("accept"), http.StatusOK)(w, r, FamilyTreeMapper(familyTree))
}

// GetListPeopleHandler godoc
// @Summary Busca todas as pessoas salvas no banco
// @Description Busca todas as pessoas salvas no banco
// @Tags person
// @Produce  json
// @Param page query int false "Página que se deseja buscar onde a página 0 é a primeira página"
// @Param size query int false "Tamanho da página"
// @Success 200 {object} GetPeopleResponse
// @Router /person [get]
func (server *Server) GetListPeopleHandler(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get(PaginationPageParam))
	if err != nil {
		page = 0
	}
	size, err := strconv.Atoi(r.URL.Query().Get(PaginationSizeParam))
	if err != nil {
		size = 0
	}

	peopleList, err := server.PersonUseCase.GetPeople(r.Context(), familytree.PaginationDetails{
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		WriteErrorMessage(w, r, http.StatusInternalServerError, ErrNotUUID)
		return
	}

	response := GetPeopleResponse{
		Content:  PeopleMapper(peopleList.Content),
		Metadata: PaginationResponseMetadata(peopleList.Metadata),
	}

	WriteJsonBody(w, r, http.StatusOK, response)
}

// PostCreatePersonHandler godoc
// @Summary Cria uma pessoa dado um body com o nome desejado
// @Description Cria uma pessoa dado um body com o nome desejado
// @Tags person
// @Produce  json
// @Param request body PostPersonRequest true "Nome da pessoa que deseja-se criar"
// @Success 201 {object} Person
// @Router /person [post]
func (server *Server) PostCreatePersonHandler(w http.ResponseWriter, r *http.Request) {
	request := &PostPersonRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	createdPerson := &familytree.Person{
		Name: request.Name,
	}
	err = server.PersonUseCase.CreatePerson(r.Context(), createdPerson)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	WriteJsonBody(w, r, http.StatusCreated, PersonMapper(*createdPerson))
}

// PostCreateParentRelationshipHandler godoc
// @Summary Cria uma relação de parentesco entre pai e filho
// @Description Cria uma relação de parentesco entre pai e filho
// @Description Não é permitido criação de relação incestuosa
// @Tags relationship
// @Produce  json
// @Param request body PostCreateParentRelationshipRequest true "Relação que deseja-se criar"
// @Success 201
// @Router /person/parent [post]
func (server *Server) PostCreateParentRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	request := &PostCreateParentRelationshipRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = request.Validate()
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = server.RelationshipUseCase.CreateParentRelation(r.Context(), request.ParentID, request.ChildID)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// DeleteParentRelationshipHandler godoc
// @Summary Remove uma relação de parentesco entre pai e filho
// @Description Remove uma relação de parentesco entre pai e filho
// @Description Não é permitido a remoção da relação se os pais do filho estiverem em uma relação de esposo e este for o único filho do casal
// @Tags relationship
// @Produce  json
// @Param request body DeleteParentRelationshipRequest true "Relação que deseja-se remover"
// @Success 204
// @Router /person/parent [delete]
func (server *Server) DeleteParentRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	request := &DeleteParentRelationshipRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = request.Validate()
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = server.RelationshipUseCase.DeleteParentRelation(r.Context(), request.ParentID, request.ChildID)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostCreateSpouseRelationshipHandler godoc
// @Summary Cria uma relação de esposo entre duas pessoas
// @Description Cria uma relação de esposo entre duas pessoas
// @Description Só é possível criar relação entre duas pessoas se elas tiverem um filho
// @Tags relationship
// @Produce  json
// @Param request body PostCreateSpouseRelationshipRequest true "Relação que deseja-se criar"
// @Success 201
// @Router /person/spouse [post]
func (server *Server) PostCreateSpouseRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	request := &PostCreateSpouseRelationshipRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = request.Validate()
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = server.RelationshipUseCase.CreateSpouseRelation(r.Context(), request.FirstSpouseID, request.SecondSpouseID)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// DeleteSpouseRelationshipHandler godoc
// @Summary Remove uma relação de esposo entre duas pessoas
// @Description Remove uma relação de esposo entre duas pessoas
// @Tags relationship
// @Produce  json
// @Param request body DeleteSpouseRelationshipRequest true "Relação que deseja-se remover"
// @Success 204
// @Router /person/spouse [delete]
func (server *Server) DeleteSpouseRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	request := &DeleteSpouseRelationshipRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = request.Validate()
	if err != nil {
		WriteErrorMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = server.RelationshipUseCase.DeleteSpouseRelation(r.Context(), request.FirstSpouseID, request.SecondSpouseID)
	if err != nil {
		WriteErrorValidation(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
