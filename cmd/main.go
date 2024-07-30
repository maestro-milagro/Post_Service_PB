package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/maestro-milagro/Post_Service_PB/internal/config"
	"github.com/maestro-milagro/Post_Service_PB/internal/http-server/handlers/delete"
	"github.com/maestro-milagro/Post_Service_PB/internal/http-server/handlers/get_all"
	"github.com/maestro-milagro/Post_Service_PB/internal/http-server/handlers/get_id"
	"github.com/maestro-milagro/Post_Service_PB/internal/http-server/handlers/post"
	"github.com/maestro-milagro/Post_Service_PB/internal/http-server/handlers/subscribe"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/service"
	"github.com/maestro-milagro/Post_Service_PB/internal/service/aws"
	"github.com/maestro-milagro/Post_Service_PB/internal/service/kafka"
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

	servicePB := service.New(log,
		storage,
		storage,
		storage,
		storage,
		storage,
		storage,
	)

	awsService := aws.New(log)

	kafkaProd := kafka.New(log, []string{cfg.KafkaBootstrapServer})

	// TODO: Метод на подписку
	router.Post("/subscribe", subscribe.New(log, servicePB))

	// TODO: Метод на пост и оповещение об этом подписчиков
	router.Post("/post", post.New(log,
		cfg.Bucket,
		cfg.Secret,
		servicePB,
		awsService,
		kafkaProd,
	))

	// TODO: Метод на вывод всех постов
	router.Get("/get_all", get_all.New(log, cfg.Secret, servicePB))

	// TODO: Метод на вывод определенного поста
	router.Get("/get_id/id={id}", get_id.New(log, cfg.Secret, cfg.Bucket, awsService, servicePB))

	// TODO: Метод на удаление поста(опцианально)
	router.Delete("/delete", delete.New(log, cfg.Secret, cfg.Bucket, awsService, servicePB))

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
