package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	"urlsh/internal/lib/api/response"
	"urlsh/internal/lib/logger/sl"
	"urlsh/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// URLGetter is an interface for getting url by alias

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
		
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("url not found for alias: ", alias)

			render.JSON(w, r, response.Error("not found"))
		
			return
		}
		if err != nil {
			log.Info("error of getting url", sl.Err(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		log.Info("got url", slog.String("url", resURL))

		// redirect to found url
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}