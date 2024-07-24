package get_all

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	Token string `json:"token"`
}

type Response struct {
	UsersInfo []models.PostUser `json:"users_info"`
	models.Response
}

type AllGetter interface {
	GetAll(ctx context.Context) ([]models.PostUser, error)
}

type CloudListDownloader interface {
	DownloadList(bucketName string) ([]types.Object, error)
}

func New(log *slog.Logger,
	secret string,
	byIDGetter AllGetter,
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
		userPost, err := byIDGetter.GetAll(r.Context())
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
		//file, err := cloud.DownloadList(bucketName)
		//if err != nil {
		//	log.Error("failed to download file", sl.Err(err))
		//
		//	render.Status(r, http.StatusInternalServerError)
		//
		//	render.JSON(w, r, models.Error("failed to download file"))
		//
		//	return
		//}

		render.JSON(w, r, Response{
			UsersInfo: userPost,
			Response:  models.OK(),
		})
	}
}
