package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/obalunenko/logger"

	"github.com/obalunenko/cthulhu-mythos-tools/internal/character"
	"github.com/obalunenko/cthulhu-mythos-tools/internal/service/assets"
	"github.com/obalunenko/cthulhu-mythos-tools/internal/storage"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mw := []func(http.Handler) http.Handler{
		logRequestMiddleware,
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

	routes := map[string]http.HandlerFunc{
		makePathPattern(http.MethodGet, "/"):                   indexHandler(),
		makePathPattern(http.MethodGet, "/favicon.ico"):        faviconHandler(),
		makePathPattern(http.MethodGet, "/characters/new"):     characterFormHandler(),
		makePathPattern(http.MethodGet, "/characters/import"):  characterImportFormHandler(),
		makePathPattern(http.MethodPost, "/characters/import"): characterImportHandler(),
		makePathPattern(http.MethodPost, "/characters"):        characterCreateHandler(),
		makePathPattern(http.MethodGet, "/characters"):         listCharactersHandler(),
		makePathPattern(http.MethodGet, "/characters/{id}"):    characterDetailsHandler(),
		makePathPattern(http.MethodDelete, "/characters/{id}"): characterDeleteHandler(),
	}

	for pattern, handler := range routes {
		logger.WithFields(context.Background(), logger.Fields{
			"endpoint": pattern,
		}).Info("Route registered")

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
			operationResponse(w, r, http.StatusInternalServerError, "Failed to render index page")

			return
		}
	}
}

func faviconHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

var charactersDB = storage.NewInMemoryStorage()

func characterFormHandler() http.HandlerFunc {
	formHTML := string(assets.MustLoad("character_create.gohtml"))
	formTmpl := template.Must(template.New("form").Parse(formHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if err := formTmpl.Execute(w, nil); err != nil {
			logger.WithError(r.Context(), err).Error("Failed to render form")

			operationResponse(w, r, http.StatusInternalServerError, "Failed to render form")
		}
	}
}

func characterImportFormHandler() http.HandlerFunc {
	formHTML := string(assets.MustLoad("character_import.gohtml"))
	formTmpl := template.Must(template.New("form").Parse(formHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if err := formTmpl.Execute(w, nil); err != nil {
			logger.WithError(r.Context(), err).Error("Failed to render form")

			operationResponse(w, r, http.StatusInternalServerError, "Failed to render form")
		}
	}
}

func characterImportHandler() http.HandlerFunc {
	const maxFileSize = 10 << 20 // Максимальный размер файла 10MB

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxFileSize); err != nil {
			operationResponse(w, r, http.StatusBadRequest, "Failed to parse form")

			return
		}

		file, _, err := r.FormFile("jsonFile")
		if err != nil {
			operationResponse(w, r, http.StatusBadRequest, "Failed to get file from form")

			logger.WithError(r.Context(), err).Error("Failed to get file from form")

			return

		}

		defer file.Close()

		var buf bytes.Buffer

		if _, err = buf.ReadFrom(file); err != nil {
			operationResponse(w, r, http.StatusBadRequest, "Failed to read file from form")

			logger.WithError(r.Context(), err).Error("Failed to read file from form")

			return
		}

		investigator, err := character.UnmarshalInvestigator(buf.Bytes())
		if err != nil {
			operationResponse(w, r, http.StatusBadRequest, "Failed to unmarshal investigator from file")

			logger.WithError(r.Context(), err).Error("Failed to unmarshal investigator from file")

			return
		}

		ch := storage.Character{
			ID:         uuid.New().String(),
			Name:       investigator.Investigator.PersonalDetails.Name,
			Occupation: investigator.Investigator.PersonalDetails.Occupation,
			Age:        investigator.Investigator.PersonalDetails.Age,
		}

		if err = charactersDB.Create(ch); err != nil {
			operationResponse(w, r, http.StatusInternalServerError, "Failed to save character to storage")

			logger.WithError(r.Context(), err).Error("Failed to save character to storage")

			return
		}

		resp := fmt.Sprintf("Character %s created!", ch.ID)

		operationResponse(w, r, http.StatusCreated, resp)
	}
}

func characterCreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Обработка данных формы
		details := storage.Character{
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
		}).Info("Create character")

		if err := charactersDB.Create(details); err != nil {
			logger.WithError(r.Context(), err).Error("Failed to save character to storage")

			operationResponse(w, r, http.StatusInternalServerError, "Failed to save character to storage")

			return
		}

		resp := fmt.Sprintf("Character %s created!", details.ID)

		operationResponse(w, r, http.StatusCreated, resp)
	}
}

func listCharactersHandler() http.HandlerFunc {
	listHTML := string(assets.MustLoad("characters.gohtml"))
	listTmpl := template.Must(template.New("list").Parse(listHTML))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		list, err := charactersDB.List()
		if err != nil {
			logger.WithError(r.Context(), err).Error("Failed to get characters list")

			operationResponse(w, r, http.StatusInternalServerError, "Failed to get characters list")

			return
		}

		if err = listTmpl.Execute(w, list); err != nil {
			logger.WithError(r.Context(), err).Error("Failed to render characters list")

			operationResponse(w, r, http.StatusInternalServerError, "Failed to render characters list")

			return
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
			resp := "Wrong character ID format"

			operationResponse(w, r, status, resp)

			return
		}

		ch, err := charactersDB.Get(id)
		if err != nil {
			var (
				status int
				resp   string
			)

			if errors.Is(err, storage.ErrNotFound) {
				status = http.StatusNotFound
				resp = "Character not found"
			} else {
				status = http.StatusInternalServerError
				resp = "Failed to get character details"
			}

			operationResponse(w, r, status, resp)

			return
		}

		if err := detailsTmpl.Execute(w, ch); err != nil {
			logger.WithError(r.Context(), err).Error("Failed to render character details")

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
			resp = "Wrong character ID format"

			return
		}

		if err := charactersDB.Delete(id); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				status = http.StatusNotFound
				resp = "Character not found"

				return
			}

			status = http.StatusInternalServerError
			resp = "Failed to delete character"

			return
		}

		status = http.StatusAccepted
		resp = fmt.Sprintf("Character %s deleted!", id)
	}
}

func isValidID(id string) bool {
	_, err := uuid.Parse(id)
	if err != nil {
		logger.WithError(context.Background(), err).
			Error("Failed to parse character ID")

		return false
	}
	return true
}

func operationResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	detailsHTML := string(assets.MustLoad("character_operation.gohtml"))
	detailsTmpl := template.Must(template.New("character_operations").Parse(detailsHTML))

	w.WriteHeader(status)

	if status != http.StatusOK && status != http.StatusCreated && status != http.StatusNoContent && status != http.StatusAccepted {
		logger.WithFields(r.Context(), logger.Fields{
			"status":  status,
			"message": message,
		}).Error("Character operation failed")

		message = fmt.Sprintf("%s[%d]: %s", http.StatusText(status), status, message)
	}

	err := detailsTmpl.Execute(w, struct {
		Message string
	}{
		Message: message,
	})
	if err != nil {
		logger.WithError(r.Context(), err).Error("Failed to render character operation response")
	}
}
