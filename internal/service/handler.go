package service

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/obalunenko/logger"

	"github.com/obalunenko/cthulhu-mythos-tools/internal/service/assets"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mw := []func(http.Handler) http.Handler{
		logRequestMiddleware,
		logResponseMiddleware,
		requestIDMiddleware,
		recoverMiddleware,
		loggerMiddleware,
	}

	mwApply := func(h http.Handler) http.Handler {
		for i := range mw {
			h = mw[i](h)
		}

		return h
	}

	type handlerWrap struct {
		http.HandlerFunc
		name string
	}

	routes := map[string]handlerWrap{
		makePathPattern(http.MethodGet, "/"): {
			HandlerFunc: indexHandler(),
			name:        "indexHandler",
		},
		makePathPattern(http.MethodGet, "/favicon.ico"): {
			HandlerFunc: faviconHandler(),
			name:        "faviconHandler",
		},
		makePathPattern(http.MethodGet, "/characters/new"): {
			HandlerFunc: characterFormHandler(),
			name:        "characterFormHandler",
		},
		makePathPattern(http.MethodPost, "/characters"): {
			HandlerFunc: characterCreateHandler(),
			name:        "characterCreateHandler",
		},
		makePathPattern(http.MethodGet, "/characters"): {
			HandlerFunc: listCharactersHandler(),
			name:        "listCharactersHandler",
		},
		makePathPattern(http.MethodGet, "/characters/{id}"): {
			HandlerFunc: characterDetailsHandler(),
			name:        "characterDetailsHandler",
		},
		makePathPattern(http.MethodDelete, "/characters/{id}"): {
			HandlerFunc: characterDeleteHandler(),
			name:        "characterDeleteHandler",
		},
	}

	for pattern, handler := range routes {
		logger.WithFields(context.Background(), logger.Fields{
			"method":  pattern,
			"handler": handler.name,
		}).Info("Добавлен обработчик")

		mux.Handle(pattern, handler)
	}

	h := http.Handler(mux)

	return mwApply(h)
}

func makePathPattern(method, path string) string {
	return fmt.Sprintf("%s %s", method, path)
}

func indexHandler() http.HandlerFunc {
	homePageHTML := string(assets.MustLoad("index.gohtml"))
	homePageTmpl := template.Must(template.New("index").Parse(homePageHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if err := homePageTmpl.Execute(w, nil); err != nil {
			http.Error(w, "failed to execute template", http.StatusInternalServerError)

			return
		}
	}
}

func faviconHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

var charactersDB = make(map[string]Character)

func characterFormHandler() http.HandlerFunc {
	formHTML := string(assets.MustLoad("character_create.gohtml"))
	formTmpl := template.Must(template.New("form").Parse(formHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if err := formTmpl.Execute(w, nil); err != nil {
			logger.WithError(r.Context(), err).Error("Ошибка при отображении формы")

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func characterCreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Обработка данных формы
		details := Character{
			ID:         uuid.New().String(),
			Name:       r.FormValue("name"),
			Occupation: r.FormValue("occupation"),
			Age:        r.FormValue("age"),
		}

		// Здесь можно добавить логику для сохранения данных персонажа
		logger.WithFields(r.Context(), logger.Fields{
			"id":         details.ID,
			"name":       details.Name,
			"occupation": details.Occupation,
			"age":        details.Age,
		}).Info("Персонаж создан")

		charactersDB[details.ID] = details

		w.WriteHeader(http.StatusCreated)

		resp := fmt.Sprintf("Персонаж %s создан!", details.ID)

		operationResponse(w, r, http.StatusCreated, resp)
	}
}

func listCharactersHandler() http.HandlerFunc {
	listHTML := string(assets.MustLoad("characters.gohtml"))
	listTmpl := template.Must(template.New("list").Parse(listHTML))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if err := listTmpl.Execute(w, charactersDB); err != nil {
			logger.WithError(r.Context(), err).Error("Ошибка при отображении списка персонажей")

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}

func characterDetailsHandler() http.HandlerFunc {
	detailsHTML := string(assets.MustLoad("character_details.gohtml"))
	detailsTmpl := template.Must(template.New("details").Parse(detailsHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		id := r.PathValue("id")
		if !isValidID(id) {
			status := http.StatusBadRequest
			resp := "Неверный формат идентификатора персонажа"

			operationResponse(w, r, status, resp)

			return
		}

		character, ok := charactersDB[id]
		if !ok {
			status := http.StatusNotFound
			resp := "Персонаж не найден"

			operationResponse(w, r, status, resp)

			return
		}

		if err := detailsTmpl.Execute(w, character); err != nil {
			logger.WithError(r.Context(), err).Error("Ошибка при отображении деталей персонажа")

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func characterDeleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			status int
			resp   string
		)

		defer func() {
			operationResponse(w, r, status, resp)
		}()

		id := r.PathValue("id")

		if !isValidID(id) {
			status = http.StatusBadRequest
			resp = "Неверный формат идентификатора персонажа"

			return
		}

		if _, ok := charactersDB[id]; !ok {
			status = http.StatusNotFound
			resp = "Персонаж не найден"
		} else {
			delete(charactersDB, id)
			status = http.StatusOK
			resp = fmt.Sprintf("Персонаж %s удален!", id)
		}
	}
}

func isValidID(id string) bool {
	_, err := uuid.Parse(id)
	if err != nil {
		logger.WithError(context.Background(), err).
			Error("Ошибка при проверке идентификатора персонажа")

		return false
	}
	return true
}

func operationResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	detailsHTML := string(assets.MustLoad("character_operation.gohtml"))
	detailsTmpl := template.Must(template.New("character_operations").Parse(detailsHTML))

	w.WriteHeader(status)

	err := detailsTmpl.Execute(w, struct {
		Message string
	}{
		Message: message,
	})
	if err != nil {
		logger.WithError(r.Context(), err).Error("Ошибка при отправке ответа")
	}
}
