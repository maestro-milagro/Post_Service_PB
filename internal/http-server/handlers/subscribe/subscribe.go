package subscribe

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"github.com/maestro-milagro/Post_Service_PB/internal/service"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type Request struct {
	models.Subscriptions
}

type Response struct {
	models.Response
}

type Subscriber interface {
	Subscribe(ctx context.Context, uid int, subId int) error
}

func New(log *slog.Logger,
	subscriber Subscriber,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.save.New"

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
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		err = subscriber.Subscribe(r.Context(), req.UID, req.SubID)
		if err != nil {
			if errors.Is(err, service.ErrSubscriptionExist) {
				log.Info("subscription already exists", slog.String("user", strconv.Itoa(req.UID)))

				render.Status(r, http.StatusBadRequest)

				render.JSON(w, r, models.Error("subscription already exists"))

				return
			}

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("error while subbing"))

			return
		}

		render.JSON(w, r, Response{
			Response: models.OK(),
		})
	}
}
