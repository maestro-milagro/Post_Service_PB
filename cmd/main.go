package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/maestro-milagro/Post_Service_PB/internal/config"
	"github.com/maestro-milagro/Post_Service_PB/internal/http-server/handlers/post"
	"github.com/maestro-milagro/Post_Service_PB/internal/http-server/handlers/subscribe"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/service"
	"github.com/maestro-milagro/Post_Service_PB/internal/service/aws"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/maestro-milagro/Post_Service_PB/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting user service",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	dbConf := postgres.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.UserName,
		Password: cfg.Password,
		DBName:   cfg.DBname,
		SSLMode:  cfg.SSLmode,
	}

	storage, err := postgres.New(dbConf)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	//	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	servicePB := service.New(log, storage, storage)

	awsService := aws.New(log)

	// TODO: Метод на подписку
	router.Post("/subscribe", subscribe.New(log, servicePB))

	// TODO: Метод на пост и оповещение об этом подписчиков
	router.Post("/post", post.New(log, cfg.Bucket, cfg.Secret, servicePB, awsService))

	// TODO: Метод на вывод всех постов

	// TODO: Метод на вывод определенного поста

	// TODO: Метод на удаление поста(опцианально)

	//router.Post("/", post.New(log, storage))
	//router.Post("/", post.New(log))
	//
	//router.Get("/email={email}&pass_hash={pass_hash}", login.New(log, storage, cfg.Secret, cfg.TokenTTL))

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")
}
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
