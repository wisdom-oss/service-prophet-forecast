package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	wisdomMiddleware "github.com/wisdom-oss/microservice-middlewares/v2"
	"microservice/globals"

	middleware2 "microservice/request/middleware"
	"microservice/request/routes"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"

	"microservice/vars"
)

/*
This function is used to set up the http server for the microservice
*/
func main() {
	vars.HttpLogger = httplog.NewLogger("prophet-forecast", httplog.Options{
		JSON:     true,
		LogLevel: "warn",
	})
	// Set up the routing of the different functions
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(httplog.RequestLogger(vars.HttpLogger))
	router.Use(wisdomMiddleware.Authorization(globals.AuthorizationConfiguration, globals.ServiceName))
	router.Use(middleware2.AdditionalResponseHeaders)
	router.Use(middleware2.ParseQueryParametersToContext)
	router.HandleFunc("/", routes.ForecastRequest)
	router.HandleFunc("/healthcheck", routes.HealthCheck)

	// Configure the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", vars.ListenPort),
		WriteTimeout: time.Second * 600,
		ReadTimeout:  time.Second * 600,
		IdleTimeout:  time.Second * 600,
		Handler:      router,
	}

	// Start the server and log errors that happen while running it
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.WithError(err).Fatal("An error occurred while starting the http server")
		}
	}()

	// Set up the signal handling to allow the server to shut down gracefully

	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, os.Interrupt)

	// Block further code execution until the shutdown signal was received
	<-cancelSignal

	log.Info("Shutting down the microservice...")

	log.Info("Closing the database connection")

}
