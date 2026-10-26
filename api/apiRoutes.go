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

	// Stage browsing
	router.HandlerFunc(http.MethodGet, "/api/v1/stages/list", api.HandleGetStageList)
	router.HandlerFunc(http.MethodPost, "/api/v1/stages/query", api.HandleGetGlobalStage)
	router.HandlerFunc(http.MethodPost, "/api/v1/stages/info", api.HandleGetStageInfo)

	// Payload submission
	router.HandlerFunc(http.MethodPost, "/api/v1/payloads/text", api.HandleUploadTextFile)
	// router.HandlerFunc(http.MethodPost, "/api/v1/payloads/file", api.HandleUploadFile)

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
