package api

import (
	"clonopher/app/burrow"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"maps"
	"net/http"
	"slices"
)

func (api *ApiConfig) HandleGetPayloadStats(w http.ResponseWriter, r *http.Request) {
	res := internal.Envelope{}

	snapShot := map[string]core.PayloadStatus{}

	core.StatsMu.RLock()
	for id, stat := range core.Stats {
		snpShot := *stat
		stat.Mu.RLock()
		for _, st := range stat.Road {
			snpShot.RoadString = append(snpShot.RoadString, st.Path)
		}
		snpShot.StageString = stat.Stage.Path
		snpShot.Stage = nil
		snpShot.Road = nil
		stat.Mu.RUnlock()
		snapShot[id] = snpShot
	}
	core.StatsMu.RUnlock()

	res["payloads"] = snapShot
	err := api.writeJSON(w, http.StatusOK, res, nil)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}
}

func (api *ApiConfig) HandleGetStageList(w http.ResponseWriter, r *http.Request) {
	res := internal.Envelope{}

	burrow.GlbsMu.RLock()
	list := slices.Collect(maps.Keys(burrow.Globals))
	burrow.GlbsMu.RUnlock()
	slices.Sort(list)
	res["list"] = list

	err := api.writeJSON(w, http.StatusOK, res, nil)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}
}

func (api *ApiConfig) HandleGetGlobalStage(w http.ResponseWriter, r *http.Request) {
	var entry struct {
		Title string `json:"select"`
	}
	err := api.readJson(w, r, &entry)
	if err != nil {
		api.badRequestResponse(w, r, err)
		return
	}
	res := internal.Envelope{}

	slected, found := burrow.GetFromGlobals(entry.Title)
	if !found {
		api.badRequestResponse(w, r, loggy.Sayf("Option %v was not found in Global Map!", entry.Title))
		return
	}
	snapShot := slected.SnapShot(10) // negative depth means unlimited shots
	res["selected"] = snapShot

	err = api.writeJSON(w, http.StatusOK, res, nil)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}
}

func (api *ApiConfig) HandleGetStageInfo(w http.ResponseWriter, r *http.Request) {
	var entry struct {
		Path string `json:"path"`
	}
	err := api.readJson(w, r, &entry)
	if err != nil {
		api.badRequestResponse(w, r, err)
		return
	}
	res := internal.Envelope{}

	stage, ok := burrow.GetFromAll(entry.Path)
	if !ok {
		res["info"] = nil
		err := api.writeJSON(w, http.StatusOK, res, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}
		return
	}

	stage.Mu.RLock()
	if stage.Info != nil {
		res["info"] = *stage.Info
	} else {
		res["info"] = nil
	}
	stage.Mu.RUnlock()

	err = api.writeJSON(w, http.StatusOK, res, nil)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}
}
