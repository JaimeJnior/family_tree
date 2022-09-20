package server

import (
	"family-tree/internal/core/familytree"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	swag "github.com/swaggo/http-swagger"
)

func NewServer(config WebConfig, router *chi.Mux, personUseCase familytree.PersonUseCasePort, relationshipUseCasePort familytree.RelationshipUseCasePort) *Server {
	return &Server{
		PersonUseCase:       personUseCase,
		RelationshipUseCase: relationshipUseCasePort,
		Router:              router,
		Config:              config,
	}
}

type Server struct {
	PersonUseCase       familytree.PersonUseCasePort
	RelationshipUseCase familytree.RelationshipUseCasePort
	Config              WebConfig
	Router              *chi.Mux
}

func (server *Server) setupMiddleware() {
	server.Router.Use(middleware.Logger)
	server.Router.Use(middleware.Recoverer)
	server.Router.Use(middleware.Timeout(time.Duration(server.Config.Timeout) * time.Second))
}

func (server *Server) setupRoutes() {

	server.Router.Get("/person", server.GetListPeopleHandler)
	server.Router.Get("/person/{personID}", server.GetPersonHandler)
	server.Router.Get("/person/{personID}/bacons/{targetPersonID}", server.GetBaconsNumber)
	server.Router.Get("/person/{personID}/tree", server.GetFamilyTree)
	server.Router.Post("/person", server.PostCreatePersonHandler)
	server.Router.Post("/person/parent", server.PostCreateParentRelationshipHandler)
	server.Router.Post("/person/spouse", server.PostCreateSpouseRelationshipHandler)
	server.Router.Delete("/person/{personID}", server.DeletePersonHandler)
	server.Router.Delete("/person/parent", server.DeleteParentRelationshipHandler)
	server.Router.Delete("/person/spouse", server.DeleteSpouseRelationshipHandler)
	server.Router.Mount("/swagger", swag.WrapHandler)

}

func (server *Server) RouteAndServe() {
	server.setupMiddleware()
	server.setupRoutes()

	http.ListenAndServe(fmt.Sprintf(":%d", server.Config.Port), server.Router)
}
