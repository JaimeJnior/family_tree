package main

import (
	"family-tree/internal/adapters/familytreerepo"
	"family-tree/internal/core/familytree"
	"family-tree/internal/server"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	"github.com/mindstand/gogm/v2"

	_ "family-tree/docs"
)

func getServerConfig() server.ServerConfig {
	cfg := &server.ServerConfig{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	if err := env.Parse(&(cfg.GogmConfig)); err != nil {
		panic(err)
	}
	if err := env.Parse(&(cfg.WebConfig)); err != nil {
		panic(err)
	}
	return *cfg
}

func setupGogm(config server.GogmConfig) *gogm.Gogm {
	gogmConfig := gogm.Config{
		Host:          config.Host,
		Port:          config.Port,
		Username:      config.Username,
		Password:      config.Password,
		PoolSize:      config.PoolSize,
		IndexStrategy: gogm.IGNORE_INDEX,
	}

	_gogm, err := gogm.New(&gogmConfig, gogm.UUIDPrimaryKeyStrategy, &familytreerepo.Person{})
	if err != nil {
		panic(err)
	}
	return _gogm
}

func setupFamilyTreeRepo(gogm *gogm.Gogm) *familytreerepo.FamilyTreeRepo {
	return familytreerepo.NewFamilyTreeRepo(gogm)
}

func setupPersonUseCase(familyTreeRepo familytree.FamilyTreeRepo) *familytree.PersonUseCase {
	return familytree.NewPersonUseCase(familyTreeRepo)
}
func setupRelationshipUseCase(familyTreeRepo familytree.FamilyTreeRepo) *familytree.RelationshipUseCase {
	return familytree.NewRelationshipUseCase(familyTreeRepo)
}

func setupServer(personUseCase familytree.PersonUseCasePort, relationShipUseCase familytree.RelationshipUseCasePort, config server.WebConfig) *server.Server {
	return server.NewServer(config, chi.NewRouter(), personUseCase, relationShipUseCase)
}

// @title Family Tree API
// @version 1.0
// @description Essa é uma api para gerenciar pessoas e relações de parentesco
// @termsOfService http://swagger.io/terms/
// @host localhost:8080
// @BasePath /
func main() {
	serverConfig := getServerConfig()
	gogm := setupGogm(serverConfig.GogmConfig)
	familyTreeRepo := setupFamilyTreeRepo(gogm)
	personUseCase := setupPersonUseCase(familyTreeRepo)
	relationShipUseCase := setupRelationshipUseCase(familyTreeRepo)
	server := setupServer(personUseCase, relationShipUseCase, serverConfig.WebConfig)
	server.RouteAndServe()

}
