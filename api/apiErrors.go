package api

import (
	"clonopher/internal"
	"clonopher/internal/loggy"
	"net/http"
)

func (api *ApiConfig) errorResponse(w http.ResponseWriter, r *http.Request, status int, AppErr error) {
	loggy.Error(AppErr)
	env := internal.Envelope{"error": AppErr.Error()}
	err := api.writeJSON(w, status, env, nil)
	if err != nil {
		w.WriteHeader(500)
	}
}

func (api *ApiConfig) serverErrorResponse(w http.ResponseWriter, r *http.Request, AppErr error) {
	api.errorResponse(w, r, http.StatusInternalServerError, AppErr)
}

func (api *ApiConfig) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	err := loggy.Say("the requested resource could not be found")
	api.errorResponse(w, r, http.StatusNotFound, err)
}

func (api *ApiConfig) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	err := loggy.Say("method does Not support this resource")
	api.errorResponse(w, r, http.StatusMethodNotAllowed, err)
}

func (api *ApiConfig) badRequestResponse(w http.ResponseWriter, r *http.Request, AppErr error) {
	api.errorResponse(w, r, http.StatusBadRequest, AppErr)
}
