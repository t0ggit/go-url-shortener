package update

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/dto"
)

type URLUpdater interface {
	UpdateURL(urlToUpdate string, alias string) error
}

func New(log *slog.Logger, urlUpdater URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.New"

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

		err = urlUpdater.UpdateURL(req.URL, req.Alias)
		if err != nil {
			log.Error("cannot update url", slog.String("error", err.Error()))
			render.JSON(w, r, dto.Response{
				Status: "ERROR",
				Error:  "cannot update url",
			})
			return
		}

		render.JSON(w, r, dto.Response{
			Status: "OK",
			Alias:  req.Alias,
		})

		log.Info("url updated", slog.String("alias", req.Alias), slog.String("url", req.URL))
	}
}
