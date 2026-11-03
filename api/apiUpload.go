package api

import (
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/fileSystem"
	"clonopher/internal/loggy"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (api *ApiConfig) HandleProcessFiles(w http.ResponseWriter, r *http.Request) {
	var entry struct {
		Name        string   `json:"name"`
		Stagings    []string `json:"stagings"` // this is gonna give us the ID for each StagedFile
		Destination string   `json:"destination"`
		Description string   `json:"description"`
	}
	err := api.readJson(w, r, &entry)
	if err != nil {
		api.badRequestResponse(w, r, err)
		return
	}
	if len(entry.Stagings) == 0 {
		api.badRequestResponse(w, r, loggy.Say("no staging file was provided!"))
		return
	}

	finalDstRoot, ok := config.Default.App.Roots[entry.Destination]
	if !ok {
		finalDstRoot = entry.Destination
	}
	ok, err = fileSystem.IsDir(finalDstRoot)
	if !ok {
		if err != nil {
			api.badRequestResponse(w, r, err)
			return
		}
		api.errorResponse(w, r, http.StatusInternalServerError, loggy.Sayf("something went wrong while validating the destination Root: %v", entry.Destination))
		return
	}

	assets, err := fileSystem.ParseAssets(entry.Stagings)
	rollback := true
	defer func() {
		if rollback {
			assets.RollbackStats()
		}
	}()
	if err != nil {
		api.errorResponse(w, r, 422, err)
		return
	}

	if entry.Name == "" {
		SetFallbackName(assets, &entry.Name)
	}

	payload, err := core.New(&config.Default, entry.Name, finalDstRoot, assets, entry.Description, internal.PlcNone, internal.ApiMode)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	err = api.writeJSON(w, http.StatusAccepted, internal.Envelope{"message": "starting to serve payload!"}, nil)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	rollback = false

	go func() {
		defer assets.RollbackStats()
		_, err := payload.Serve()
		if err != nil {
			payload.ShiftError(err)
			return
		}

		err = payload.Drop()
		if err != nil {
			payload.ShiftError(err)
			return
		}
	}()
}

func (api *ApiConfig) HandleProcessTextFile(w http.ResponseWriter, r *http.Request) {
	var entry struct {
		FileName    string `json:"fileName"`
		Head        string `json:"head"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}
	err := api.readJson(w, r, &entry)
	if err != nil {
		api.badRequestResponse(w, r, err)
		return
	}

	entry.FileName = strings.TrimSpace(entry.FileName)
	if strings.ContainsAny(entry.FileName, "/\\") || entry.FileName == "." || entry.FileName == ".." || entry.FileName == "" {
		api.badRequestResponse(w, r, loggy.Say("file-name should not be empty or containing path charectors like \\ / .."))
		return
	}

	payload, err := core.New(&config.Default, entry.FileName, entry.Head, []*fileSystem.Asset{}, entry.Description, internal.PlcNone, internal.ApiMode)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	err = api.writeJSON(w, http.StatusAccepted, internal.Envelope{"message": "starting to serve payload!"}, nil)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	go func() {
		dst, err := payload.Serve()
		if err != nil {
			payload.ShiftError(err)
			return
		}
		err = payload.Dir.Submit()
		if err != nil {
			payload.ShiftError(err)
			return
		}

		path := filepath.Join(dst, entry.FileName)

		err = os.WriteFile(path, []byte(entry.Content), 0644)
		if err != nil {
			payload.ShiftError(err)
			return
		}
	}()
}

func (api *ApiConfig) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.methodNotAllowedResponse(w, r)
		return
	}

	const maxUploadSize = 100 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		api.badRequestResponse(w, r, loggy.Say("Invalid multipart form or file too large"))
		return
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	files := r.MultipartForm.File["file"]
	names := r.MultipartForm.Value["filename"]

	if len(files) != 1 || len(names) != 1 {
		api.badRequestResponse(w, r, loggy.Say("Exactly one file and filename are required"))
		return
	}

	filename := strings.TrimSpace(names[0])
	filename = filepath.Base(strings.ReplaceAll(filename, `\\`, `/`))

	if filename == "" || filename == "." || filename == ".." {
		api.badRequestResponse(w, r, loggy.Say("Invalid filename"))
		return
	}

	src, err := files[0].Open()
	if err != nil {
		api.serverErrorResponse(w, r, loggy.Say("Cannot open uploaded file"))
		return
	}
	defer src.Close()

	uploadDir := config.Default.App.DefaultUploadDir

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		api.serverErrorResponse(w, r, loggy.Say("Cannot create upload directory"))
		return
	}

	dst, actualName, err := createUniqueUploadFile(uploadDir, filename)
	if err != nil {
		api.serverErrorResponse(w, r, loggy.Say("Cannot create destination file"))
		return
	}

	size, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()

	if copyErr != nil || closeErr != nil {
		os.Remove(filepath.Join(uploadDir, actualName))
		api.serverErrorResponse(w, r, loggy.Say("Cannot save uploaded file"))
		return
	}

	err = fileSystem.GlobalStagings.AddNewStaging(actualName, filepath.Join(uploadDir, actualName))
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	res := internal.Envelope{
		"name": actualName,
		"size": size,
	}

	api.writeJSON(w, http.StatusCreated, res, nil)
}

func createUniqueUploadFile(dir, filename string) (*os.File, string, error) {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	for i := 0; i < 10000; i++ {
		candidate := filename

		if i > 0 {
			candidate = fmt.Sprintf("%s (%d)%s", name, i, ext)
		}

		path := filepath.Join(dir, candidate)

		file, err := os.OpenFile(
			path,
			os.O_WRONLY|os.O_CREATE|os.O_EXCL,
			0644,
		)

		if err == nil {
			return file, candidate, nil
		}

		if !errors.Is(err, os.ErrExist) {
			return nil, "", err
		}
	}

	return nil, "", errors.New("too many filename collisions")
}
