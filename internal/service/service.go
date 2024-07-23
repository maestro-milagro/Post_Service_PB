package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"github.com/maestro-milagro/Post_Service_PB/internal/storage"
	"log/slog"
	"strconv"
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
	dbByIDGetter DBByIDGetter
}

func New(log *slog.Logger, dbSubscriber DBSubscriber, dbPostSaver DBPostSaver, dbByIDGetter DBByIDGetter) *Service {
	return &Service{
		log:          log,
		dbSubscriber: dbSubscriber,
		dbPostSaver:  dbPostSaver,
		dbByIDGetter: dbByIDGetter,
	}
}

type DBSubscriber interface {
	SubscribeDB(ctx context.Context, uid int, subId int) error
}

type DBPostSaver interface {
	PostSaveDB(ctx context.Context, user models.PostUser) (int64, error)
}

type DBByIDGetter interface {
	GetByIdDB(ctx context.Context, id int) (models.PostUser, error)
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

func (s *Service) GetById(ctx context.Context, id int) (models.PostUser, error) {
	const op = "service.GetById"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", strconv.Itoa(id)),
	)

	log.Info("getting by id")

	userPost, err := s.dbByIDGetter.GetByIdDB(ctx, id)
	if err != nil {
		log.Error("error while getting by id", sl.Err(err))

		return models.PostUser{}, fmt.Errorf("%s: %w", op, err)
	}
	return userPost, nil
}
