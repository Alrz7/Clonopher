package api

import (
	"clonopher/internal"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (api *ApiConfig) routes() *httprouter.Router {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/", api.handleHome)

	// Payload monitoring
	router.HandlerFunc(http.MethodGet, "/api/v1/payloads/stats", api.HandleGetPayloadStats)
	router.HandlerFunc(http.MethodGet, "/api/v1/logs", api.HandleGetGlobalLogs)
	router.HandlerFunc(http.MethodGet, "/api/v1/stagings", api.HandleGetStagingFiles)

	// Directory browsing
	router.HandlerFunc(http.MethodGet, "/api/v1/directories/list", api.HandleGetDirList)
	router.HandlerFunc(http.MethodPost, "/api/v1/directories/query", api.HandleGetGlobalDirectory)
	router.HandlerFunc(http.MethodPost, "/api/v1/directories/info", api.HandleGetDirInfo)

	// Payload submission
	router.HandlerFunc(http.MethodPost, "/api/v1/payloads/new", api.HandleProcessFiles)
	router.HandlerFunc(http.MethodPost, "/api/v1/payloads/text", api.HandleProcessTextFile)
	router.HandlerFunc(http.MethodPost, "/api/v1/payloads/file", api.HandleUploadFile)
	router.HandlerFunc(http.MethodDelete, "/api/v1/payloads/file", api.HandleDeleteStaging)

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
