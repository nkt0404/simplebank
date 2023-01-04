package main

import (
	"database/sql"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"

	//_ "github.com/nkt0404//simplebank/doc/statik"
	"github.com/nkt0404/simplebank/api"
	db "github.com/nkt0404/simplebank/db/sqlc"
	"github.com/nkt0404/simplebank/util"

	//"github.com/nkt0404/simplebank/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// const (
// 	dbDriver      = "postgres"
// 	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
// 	serverAddress = "0.0.0.0:8080"
// )

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	//go runTaskProcessor(redisOpt, store)
	//go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}

// func runTaskProcessor(redisOpt asynq.RedisClientOpt, store *db.Store) {
// 	processor := worker.NewTaskProcessor(redisOpt, store)
// 	log.Info().Msg("start task processor")
// 	err := processor.Start()
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to start task processor")
// 	}
// }

// func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
// 	server, err := gapi.NewServer(config, store, taskDistributor)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("cannot create server")
// 	}
// 	err = server.Start(config.GRPCServerAddress)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("cannot start grpc server")
// 	}
// }

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
}
