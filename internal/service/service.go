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
	ErrUserNotFound      = errors.New("no users found")
	ErrSubscriptionExist = errors.New("subscription already exists")
	ErrNoFollowers       = errors.New("no followers found")
)

type Service struct {
	log          *slog.Logger
	dbSubscriber DBSubscriber
	dbPostSaver  DBPostSaver
	dbByIDGetter DBByIDGetter
	dbAllGetter  DBAllGetter
	dbDeleter    DBDeleter
	dbWhoSubbed  DBWhoSubbed
}

func New(log *slog.Logger,
	dbSubscriber DBSubscriber,
	dbPostSaver DBPostSaver,
	dbByIDGetter DBByIDGetter,
	dbAllGetter DBAllGetter,
	dbDeleter DBDeleter,
	dbWhoSubbed DBWhoSubbed) *Service {
	return &Service{
		log:          log,
		dbSubscriber: dbSubscriber,
		dbPostSaver:  dbPostSaver,
		dbByIDGetter: dbByIDGetter,
		dbAllGetter:  dbAllGetter,
		dbDeleter:    dbDeleter,
		dbWhoSubbed:  dbWhoSubbed,
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

type DBAllGetter interface {
	GetAllDB(ctx context.Context) ([]models.PostUser, error)
}

type DBDeleter interface {
	DeleteDB(ctx context.Context, ids []int) ([]string, error)
}

type DBWhoSubbed interface {
	WhoSubbedDB(ctx context.Context, email string) ([]int, error)
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

func (s *Service) GetAll(ctx context.Context) ([]models.PostUser, error) {
	const op = "service.GetAll"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("getting all")

	userPost, err := s.dbAllGetter.GetAllDB(ctx)
	if err != nil {
		log.Error("error while getting all", sl.Err(err))

		return []models.PostUser{}, fmt.Errorf("%s: %w", op, err)
	}
	if len(userPost) == 0 {
		return []models.PostUser{}, fmt.Errorf("%s: %w", op, ErrNoFollowers)
	}
	return userPost, nil
}

func (s *Service) Delete(ctx context.Context, ids []int) ([]string, error) {
	const op = "service.GetById"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("deleting by id")

	delObjects, err := s.dbDeleter.DeleteDB(ctx, ids)
	if err != nil {
		log.Error("error while getting by id", sl.Err(err))

		return []string{}, fmt.Errorf("%s: %w", op, err)
	}
	return delObjects, nil
}

func (s *Service) WhoSubbed(ctx context.Context, email string) ([]int, error) {
	const op = "service.WhoSubbedDB"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("searching for subbs")

	subbs, err := s.dbWhoSubbed.WhoSubbedDB(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNoFollowers) {
			log.Error("no followers were found", sl.Err(err))

			return []int{}, ErrNoFollowers
		}
		log.Error("error while searching for subbs", sl.Err(err))

		return []int{}, fmt.Errorf("%s: %w", op, err)
	}
	return subbs, nil
}
