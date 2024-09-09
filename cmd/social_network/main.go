package main

import (
	"context"
	"os"
	"os/signal"
	"social_network/internal/controllers"
	"social_network/internal/repositories"
	"social_network/internal/servers"
	"social_network/internal/services"
	"sync"
	"syscall"
)

const (
	UrlDb                = "urlDb"
	DefaultAddressServer = ":8080"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	urlDb := getEnv("DB_URL", UrlDb)
	addressServer := getEnv("ADDRESS_SERVER", DefaultAddressServer)

	logger := initLogger()
	database := initDatabase(urlDb, logger)

	profileRepository := repositories.NewProfileRepository(database, logger)
	profileService := services.NewProfileService(&profileRepository)
	profileController := controllers.NewProfileController(&profileService, logger)

	restServer := servers.NewRESTServer(addressServer, nil, &profileController, logger)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		restServer.ListenAndServe()
	}()

	go func() {
		defer wg.Done()
		<-ctx.Done()

		restServer.Shutdown(context.Background())
		logger.Info("REST server closed")

		_ = database.Close()
		logger.Info("database closed")
	}()

	wg.Wait()
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
