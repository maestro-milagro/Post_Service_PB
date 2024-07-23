package get_id

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/jwt"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"github.com/maestro-milagro/Post_Service_PB/internal/storage"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type Request struct {
	Token string `json:"token"`
}

type Response struct {
	models.Response
}

type ByIDGetter interface {
	GetById(ctx context.Context, id int) (models.PostUser, error)
}

type CloudDownloader interface {
	DownloadFile(bucketName string, filename string) ([]byte, error)
}

func New(log *slog.Logger,
	secret string,
	bucketName string,
	cloud CloudDownloader,
	byIDGetter ByIDGetter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get_id.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("empty request"))
			return
		}
		_, err = jwt.VerifyToken(log, secret, req.Token)
		if err != nil {
			log.Error("failed to verify token", sl.Err(err))

			render.Status(r, http.StatusUnauthorized)

			render.JSON(w, r, models.Error("invalid token"))
		}

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("invalid request"))

			return
		}
		newId, err := strconv.Atoi(id)
		if err != nil {
			log.Error("failed to parse id", sl.Err(err))
		}
		userPost, err := byIDGetter.GetById(r.Context(), newId)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Warn("user not found", sl.Err(err))

				render.Status(r, http.StatusNotFound)

				render.JSON(w, r, models.Error("invalid credentials"))

				return
			}
			log.Error("failed to get user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to get user"))

			return
		}

		// TODO: aws download
		somesh, err := cloud.DownloadFile(bucketName, userPost.Key)

		render.JSON(w, r, Response{
			Response: models.OK(),
		})
	}
}
