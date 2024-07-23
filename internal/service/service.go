package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"github.com/maestro-milagro/Post_Service_PB/internal/storage"
	"log/slog"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSubscriptionExist  = errors.New("subscription already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	log          *slog.Logger
	dbSubscriber DBSubscriber
	dbPostSaver  DBPostSaver
}

func New(log *slog.Logger, dbSubscriber DBSubscriber, dbPostSaver DBPostSaver) *Service {
	return &Service{
		log:          log,
		dbSubscriber: dbSubscriber,
		dbPostSaver:  dbPostSaver,
	}
}

type DBSubscriber interface {
	SubscribeDB(ctx context.Context, uid int, subId int) error
}

type DBPostSaver interface {
	PostSaveDB(ctx context.Context, user models.PostUser) (int64, error)
}

func (s *Service) Subscribe(ctx context.Context, uid int, subId int) error {
	const op = "service.Subscribe"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("userID", int64(uid)),
	)

	log.Info("subscribing user")

	err := s.dbSubscriber.SubscribeDB(ctx, uid, subId)
	if err != nil {
		if errors.Is(err, storage.ErrSubExist) {
			s.log.Warn("already following", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrSubscriptionExist)
		}

		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Service) SavePost(ctx context.Context, user models.PostUser) (int64, error) {
	const op = "service.SavePost"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user", user.Email),
	)

	log.Info("saving post")

	id, err := s.dbPostSaver.PostSaveDB(ctx, user)
	if err != nil {
		log.Error("error while saving post", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}
