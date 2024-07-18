package post

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/jwt"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"io"
	"log/slog"
	"net/http"
)

// TODO: mb

type Request struct {
	Token string `json:"token"`
}

type Response struct {
	models.Response
}

type PostUserSaver interface {
	SavePost(ctx context.Context, user models.PostUser) (int64, error)
}

func New(log *slog.Logger,
	secret string,
	postUserSaver PostUserSaver,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		//var req Request
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("empty request"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		email, err := jwt.VerifyToken(log, secret, req.Token)
		if err != nil {
			log.Error("failed to verify token", sl.Err(err))

			render.Status(r, http.StatusUnauthorized)

			render.JSON(w, r, models.Error("invalid token"))
		}

		//TODO: AWS integration
		bucket := ""
		key := ""

		id, err := postUserSaver.SavePost(r.Context(), models.PostUser{
			Email:  email,
			Bucket: bucket,
			Key:    key,
		})

		// TODO: id is given to notification service

		render.JSON(w, r, Response{
			Response: models.OK(),
		})
	}
}
