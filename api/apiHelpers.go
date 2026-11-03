package api

import (
	"clonopher/internal"
	"clonopher/internal/fileSystem"
	"clonopher/internal/loggy"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (api *ApiConfig) writeJSON(w http.ResponseWriter, status int, data internal.Envelope, headers http.Header) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	maps.Copy(w.Header(), headers)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
	return nil
}

// dst must be a pointer to the destination
func (api *ApiConfig) readJson(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return loggy.Sayf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return loggy.Sayf("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return loggy.Sayf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return loggy.Sayf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return loggy.Say("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return loggy.Sayf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return loggy.Sayf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return loggy.Say("body must only contain a single JSON value")
	}

	return nil
}

func SetFallbackName(files []*fileSystem.Asset, name *string) {
	var names []string

	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f.FileName), filepath.Ext(f.FileName))
		name = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))

		if name == "" {
			continue
		}

		if r := []rune(name); len(r) > 8 {
			name = string(r[:8])
		}

		names = append(names, name)
		if len(names) == 2 {
			break
		}
	}

	if len(names) == 0 {
		names = append(names, "payload")
	}

	now := time.Now()
	*name = strings.Join(names, "-") + "-" +
		now.Format("060102-150405") + "-" +
		strconv.FormatInt(int64(now.Nanosecond()), 36)
}
