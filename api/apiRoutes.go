package api

import (
	"clonopher/internal"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (api *ApiConfig) routes() *httprouter.Router {

	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/", api.handleHome)
	router.HandlerFunc(http.MethodPost, "/payload", api.HandleUploadTextFile)
	return router
}

func (api *ApiConfig) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		api.notFoundResponse(w, r)
		return
	}
	api.HealthStatus(w, r)
}

func (api *ApiConfig) HealthStatus(w http.ResponseWriter, r *http.Request) {
	stat := internal.Envelope{
		"status": "available",
	}
	err := api.writeJSON(w, http.StatusOK, stat, nil)
	if err != nil {
		api.serverErrorResponse(w, r, err)
	}
}
