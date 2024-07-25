package delete

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/jwt"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"github.com/maestro-milagro/Post_Service_PB/internal/storage"
	"io"
	"log/slog"
	"net/http"
)

type Request struct {
	IDs   []int  `json:"ids"`
	Token string `json:"token"`
}

type Response struct {
	models.Response
}

type Deleter interface {
	Delete(ctx context.Context, ids []int) ([]string, error)
}

type CloudDeleter interface {
	DeleteObjects(bucketName string, objectKeys []string) error
}

func New(log *slog.Logger,
	secret string,
	bucketName string,
	cloud CloudDeleter,
	deleter Deleter,
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

			return
		}

		//id := chi.URLParam(r, "id")
		//if id == "" {
		//	log.Info("id is empty")
		//
		//	render.Status(r, http.StatusBadRequest)
		//
		//	render.JSON(w, r, models.Error("invalid request"))
		//
		//	return
		//}
		if err != nil {
			log.Error("failed to parse id", sl.Err(err))
		}
		delObjects, err := deleter.Delete(r.Context(), req.IDs)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Warn("users not found", sl.Err(err))

				render.Status(r, http.StatusNotFound)

				render.JSON(w, r, models.Error("invalid credentials"))

				return
			}
			log.Error("failed to delete users", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to delete user"))

			return
		}

		// TODO: aws download
		err = cloud.DeleteObjects(bucketName, delObjects)
		if err != nil {
			log.Error("failed to download file", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to download file"))

			return
		}

		render.JSON(w, r, Response{
			Response: models.OK(),
		})
	}
}
