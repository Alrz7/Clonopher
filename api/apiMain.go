package api

import (
	"clonopher/app/config"
	"clonopher/internal/loggy"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ApiConfig struct {
	Ctx     context.Context
	Close   context.CancelFunc
	Options *config.Config
}

func (api *ApiConfig) Run() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", api.Options.Api.Port),
		Handler:      api.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	loggy.Info("Starting clonopher on", "url", fmt.Sprintf("http://localhost:%d/", api.Options.Api.Port))

	err := srv.ListenAndServe()

	switch {
	case errors.Is(err, http.ErrServerClosed):
		// loggy.Info(err)
	default:
		loggy.Fatal("There was an error while starting the Api server", "err", err.Error())
	}
}
