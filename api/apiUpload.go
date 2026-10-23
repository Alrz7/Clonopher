package api

import (
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (api *ApiConfig) HandleUploadTextFile(w http.ResponseWriter, r *http.Request) {
	var entry struct {
		FileName    string `json:"fileName"`
		Head        string `json:"head"`
		Content     string `json:"content"`
		Description string `json:"description"`
		Policy      string `json:"policy"`
	}
	err := api.readJson(w, r, &entry)
	if err != nil {
		api.badRequestResponse(w, r, err)
		return
	}

	payload, err := core.New(&config.Default, entry.FileName, entry.Head, "", entry.Description, internal.PlcNone, internal.ApiMode)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}
	dst, err := payload.Serve()
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}
	err = payload.Stage.Submit()
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	path := filepath.Join(dst, entry.FileName)

	err = os.WriteFile(path, []byte(entry.Content), 0644)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	err = api.writeJSON(w, http.StatusOK, internal.Envelope{"message": "payload served!", "address": path}, nil)
	if err != nil {
		loggy.Error(err)
	}
}

const (
	maxUploadSize      int64 = 100 << 20 // 100 MiB
	maxMultipartMemory int64 = 8 << 20   // 8 MiB
	maxRequestOverhead int64 = 1 << 20   // 1 MiB
	maxDescriptionSize       = 64 << 10  // 64 KiB
)

type uploadDraft struct {
	ID           string `json:"id"`
	OriginalName string `json:"original_name"`
	Description  string `json:"description"`
	Size         int64  `json:"size"`
	Status       string `json:"status"`
}

func UploadHandler(uploadDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(
			w, r.Body, maxUploadSize+maxRequestOverhead,
		)

		if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}

			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(
					w, "request body too large",
					http.StatusRequestEntityTooLarge,
				)
				return
			}

			http.Error(
				w, "invalid multipart request",
				http.StatusBadRequest,
			)
			return
		}

		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}

		description := strings.TrimSpace(r.FormValue("description"))
		if len(description) > maxDescriptionSize {
			http.Error(
				w, "description too large",
				http.StatusRequestEntityTooLarge,
			)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(
				w, "missing file field",
				http.StatusBadRequest,
			)
			return
		}
		defer file.Close()

		if err := os.MkdirAll(uploadDir, 0700); err != nil {
			http.Error(
				w, "storage unavailable",
				http.StatusInternalServerError,
			)
			return
		}

		dst, err := os.CreateTemp(uploadDir, "payload-*")
		if err != nil {
			http.Error(
				w, "cannot create upload",
				http.StatusInternalServerError,
			)
			return
		}

		stagedPath := dst.Name()
		id := filepath.Base(stagedPath)
		metadataPath := filepath.Join(uploadDir, id+".json")

		keep := false
		defer func() {
			if !keep {
				_ = os.Remove(stagedPath)
				_ = os.Remove(metadataPath)
			}
		}()

		size, copyErr := io.Copy(
			dst, io.LimitReader(file, maxUploadSize+1),
		)
		closeErr := dst.Close()

		if copyErr != nil || closeErr != nil {
			http.Error(
				w, "failed to save upload",
				http.StatusInternalServerError,
			)
			return
		}

		if size > maxUploadSize {
			http.Error(
				w, "file too large",
				http.StatusRequestEntityTooLarge,
			)
			return
		}

		draft := uploadDraft{
			ID:           id,
			OriginalName: header.Filename,
			Description:  description,
			Size:         size,
			Status:       "uploaded",
		}

		metadata, err := json.MarshalIndent(draft, "", "  ")
		if err != nil {
			http.Error(
				w, "cannot encode metadata",
				http.StatusInternalServerError,
			)
			return
		}

		if err := os.WriteFile(metadataPath, metadata, 0600); err != nil {
			http.Error(
				w, "cannot save metadata",
				http.StatusInternalServerError,
			)
			return
		}

		keep = true

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(draft)
	}
}
