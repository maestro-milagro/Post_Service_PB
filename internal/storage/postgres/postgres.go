package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"github.com/maestro-milagro/Post_Service_PB/internal/storage"
)

type Storage struct {
	db *sqlx.DB
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func New(cfg Config) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SubscribeDB(ctx context.Context, uid int, subId int) error {
	const op = "storage.postgres.NewSubscribeDB"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	createListQuery := fmt.Sprintf("INSERT INTO subscriptions (uid, sub_id) VALUES ($1, $2)")

	_, err = tx.Exec(createListQuery, uid, subId)
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			switch e.Code {
			case "23505":
				// p-key constraint violation
				tx.Rollback()
				return fmt.Errorf("%s: %w", op, storage.ErrSubExist)
			default:
				tx.Rollback()
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}
	return tx.Commit()
}

func (s *Storage) PostSaveDB(ctx context.Context, user models.PostUser) (int64, error) {
	const op = "Storage/postgres/SaveUser"

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int
	createListQuery := fmt.Sprintf("INSERT INTO users_posts (email, bucket, key) VALUES ($1, $2, $3) RETURNING id")
	row := tx.QueryRow(createListQuery, user.Email, user.Bucket, user.Key)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int64(id), tx.Commit()
}

func (s *Storage) GetByIdDB(ctx context.Context, id int) (models.PostUser, error) {
	const op = "Storage/postgres/User"

	var user models.PostUser

	createListQuery := fmt.Sprintf("SELECT users_posts.email, users_posts.bucket, users_posts.key FROM users_posts WHERE id = $1")

	row := s.db.QueryRow(createListQuery, id)

	err := row.Scan(&user.Email, &user.Bucket, &user.Key)
	if err != nil {
		return models.PostUser{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
