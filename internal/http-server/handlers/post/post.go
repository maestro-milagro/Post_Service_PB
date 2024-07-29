package post

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/jwt"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"io"
	"log/slog"
	"net/http"
)

type Request struct {
	Token string `json:"token"`
}

type Response struct {
	Id int `json:"id"`
	models.Response
}

type PostUserSaver interface {
	SavePost(ctx context.Context, user models.PostUser) (int64, error)
}

type CloudUploader interface {
	UploadFile(bucketName string, fileName string, largeObject []byte) error
}

type Producer interface {
	Produce(post models.Post, topic string)
}

func New(log *slog.Logger,
	bucket string,
	secret string,
	postUserSaver PostUserSaver,
	uploader CloudUploader,
	producer Producer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := r.ParseMultipartForm(100)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to decode request"))

			return
		}
		mForm := r.MultipartForm

		req.Token = mForm.Value["token"][0]
		if req.Token == "" {
			log.Error("token is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("empty token"))
			return
		}

		var key string
		var fileObject []byte

		for k, _ := range mForm.File {
			// k is the key of file part
			file, fileHeader, err := r.FormFile(k)
			if err != nil {
				log.Error("invoke FormFile error:")

				render.Status(r, http.StatusBadRequest)

				render.JSON(w, r, models.Error("bed request"))

				return
			}
			key = fileHeader.Filename
			fileObject, err = io.ReadAll(file)
			if err != nil {
				log.Error("invoke ReadAll error:")

				render.Status(r, http.StatusInternalServerError)

				render.JSON(w, r, models.Error("error while reading file"))
			}
		}

		log.Info("request body decoded", slog.Any("request", req))

		email, err := jwt.VerifyToken(log, secret, req.Token)
		if err != nil {
			log.Error("failed to verify token", sl.Err(err))

			render.Status(r, http.StatusUnauthorized)

			render.JSON(w, r, models.Error("invalid token"))
		}

		//TODO: AWS integration
		err = uploader.UploadFile(bucket, key, fileObject)
		if err != nil {
			log.Error("failed to upload file", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to upload file"))
		}

		fmt.Println(fileObject)

		id, err := postUserSaver.SavePost(r.Context(), models.PostUser{
			Email:  email,
			Bucket: bucket,
			Key:    key,
		})
		if err != nil {
			log.Error("failed to save post", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to save post"))
		}

		// TODO: notification service
		producer.Produce(models.Post{PostID: int(id), Email: email}, "posts")

		render.JSON(w, r, Response{
			Id:       int(id),
			Response: models.OK(),
		})
	}
}
