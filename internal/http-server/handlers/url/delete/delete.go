package delete

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/dto"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req dto.Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("cannot decode request body", slog.String("error", err.Error()))
			render.JSON(w, r, dto.Response{
				Status: "ERROR",
				Error:  "cannot decode request body",
			})
			return
		}

		alias := req.Alias

		err = urlDeleter.DeleteURL(alias)
		if err != nil {
			log.Error("cannot delete url", slog.String("error", err.Error()))
			render.JSON(w, r, dto.Response{
				Status: "ERROR",
				Error:  "cannot delete url",
			})
			return
		}

		log.Info("url deleted")

		render.JSON(w, r, dto.Response{
			Status: "OK",
			Alias:  alias,
		})
		return
	}
}
