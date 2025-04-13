package save

import (
	"errors"
	"log/slog"
	"net/http"
	"urlsh/internal/lib/api/response"
	"urlsh/internal/lib/logger/sl"
	"urlsh/internal/lib/random"
	"urlsh/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias  string `json:"alias,omitempty"`
}

// TODO: move to config
const aliasLength = 6

//go:generate moq -out URLSaver.go . URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("rquest_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))
	
		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid of fields of request", sl.Err(err))

			render.JSON(w, r, response.Error("invalid of fields of request"))

			return
		} 

		alias := req.Alias
		if alias == "" {
			alias = random.RandomAlias(aliasLength)
		}
		// TODO: create func for check alias with other aliases already used
		
		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, response.Error("url already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, response.Error("failed to add url"))
		
			return
		}

		log.Info("url added", slog.Int64("id", id))

		// 
		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias: alias,
		})
	}
}