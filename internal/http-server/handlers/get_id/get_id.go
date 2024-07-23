package get_id

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"log/slog"
	"net/http"
)

type Request struct {
	Token string `json:"token"`
}

type Response struct {
	models.Response
}

type PostUserSaver interface {
	SavePost(ctx context.Context, user models.PostUser) (int64, error)
}

type CloudDownloader interface {
	DownloadFile(bucketName string, filename string) ([]byte, error)
}

func New(log *slog.Logger,
	bucket string,
	secret string,
	cloud CloudDownloader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get_id.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("invalid request"))

			return
		}

		render.JSON(w, r, Response{
			Response: models.OK(),
		})
	}
}
