package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/EM-Stawberry/Stawberry/pkg/email"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/gin-gonic/gin"
)

func StartServer(
	router *gin.Engine,
	mailer email.MailerService,
	cfg *config.ServerConfig,
	log *zap.Logger) error {
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	switch cfg.GinMode {
	case gin.DebugMode:
		gin.SetMode(gin.DebugMode)
	case gin.TestMode:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("error starting server: %w", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info("Initiating shutdown", zap.Any("signal", sig))

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		mailer.Stop(ctx)

		if err := srv.Shutdown(ctx); err != nil {
			_ = srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
