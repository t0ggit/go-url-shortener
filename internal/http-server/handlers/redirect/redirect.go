package redirect

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/dto"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias is empty")
			render.JSON(w, r, dto.Response{
				Status: "ERROR",
				Error:  "alias is empty",
			})
			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if err != nil {
			log.Error("cannot get url", slog.String("error", err.Error()))
			render.JSON(w, r, dto.Response{
				Status: "ERROR",
				Error:  "cannot get url",
			})
			return
		}

		log.Info("redirect to url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
