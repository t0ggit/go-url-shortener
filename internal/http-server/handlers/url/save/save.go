package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/dto"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

const (
	randomAliasLength = 7
)

type URLSaver interface {
	SaveURL(urlToSave string, alias string) error
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

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

		log.Info("request body decoded", slog.Any("request", req))

		err = validator.New().Struct(req)
		if err != nil {
			log.Error("invalid request body", slog.String("error", err.Error()))

			render.JSON(w, r, dto.Response{
				Status: "ERROR",
				Error:  "invalid request body",
			})
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(randomAliasLength)
		}

		err = urlSaver.SaveURL(req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				log.Error("url already exists", slog.String("error", err.Error()))

				render.JSON(w, r, dto.Response{
					Status: "ERROR",
					Error:  "url already exists",
				})
				return
			}

			log.Error("cannot save url", slog.String("error", err.Error()))

			render.JSON(w, r, dto.Response{
				Status: "ERROR",
				Error:  "cannot save url",
			})
			return
		}

		log.Info("url saved", slog.String("alias", alias))

		render.JSON(w, r, dto.Response{
			Status: "OK",
			Alias:  alias,
		})
		return
	}
}
