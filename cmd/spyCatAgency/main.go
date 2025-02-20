package main

import (
	"spyCatAgency/internal/domain/usecases"
	"spyCatAgency/internal/infrastructure/database"
	"spyCatAgency/internal/infrastructure/logger"
	"spyCatAgency/internal/presentation/server"
	"spyCatAgency/internal/presentation/server/handlers"

	"github.com/spf13/viper"
)

func main() {
	logger := logger.NewZapLogger()
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal("Failed to read config file:", err)
	}
	dbSource := viper.GetString("DB_SOURCE")
	db := database.Init(dbSource)

	migrationPath := viper.GetString("MIGRATION_PATH")
	database.RunDBMigration(migrationPath, dbSource)

	catRepo := database.NewCatRepository(logger, db)
	catUseCase := usecases.NewCatUseCase(logger, catRepo)
	catHandler := handlers.NewCatHandler(logger, catUseCase)

	missionRepo := database.NewMissonRepository(logger, db)
	missionUseCase := usecases.NewMissionUseCase(logger, missionRepo, catRepo)
	missionHandler := handlers.NewMisionHandler(logger, missionUseCase)

	app := server.New(logger, catHandler, missionHandler)

	port := viper.GetString("SERVER_PORT")
	app.Run(port)

}
