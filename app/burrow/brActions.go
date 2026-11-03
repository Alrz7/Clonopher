package burrow

import (
	"clonopher/app/providers"
	"clonopher/internal"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/spf13/fileflow"
	"go.yaml.in/yaml/v3"
)

func (d *Directory) ReadHelperFile() error {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	return d.readHelperFileLocked()
}

func (d *Directory) readHelperFileLocked() error {
	if d.Info != nil {
		return nil
	}

	if fileflow.Exists(filepath.Join(d.Path, internal.HelperFileName)) {
		b, err := os.ReadFile(filepath.Join(d.Path, internal.HelperFileName))
		if err != nil {
			return err
		}
		var inf DirInfo
		err = yaml.Unmarshal(b, &inf)
		if err != nil {
			return err
		}

		if filepath.Clean(inf.Path) != filepath.Clean(d.Path) {
			loggy.Warnf("helper file's path <%v> \n does NOT match the Directory its placed in <%v> \n", inf.Path, d.Path)
		}
		d.setInfoLocked(&inf)
	}
	return nil
}

func (d *Directory) InferDirInfo(provider *providers.Provider) error {
	d.Mu.RLock()
	if d.Info != nil {
		d.Mu.RUnlock()
		return nil
	}

	loggy.AddNew("generating Info...", loggy.PdLInfo)
	feed := helpers.InfoOptions{
		&helpers.InfoOption{Info: `dir-path: %v`, Options: []any{d.Path}},
		&helpers.InfoOption{Info: `sub-dirs: %v`, Options: []any{slices.Collect(maps.Keys(d.Children))}},
		&helpers.InfoOption{Info: `dir-files: %v`, Options: []any{d.Files}},
	}.ConcatIfExists()

	d.Mu.RUnlock()

	query := `You are Clonopher's directory metadata generator.

Your task is to produce a concise, factual brief describing the
current directory for use by a separate file-classification model.

Rules:
- Output only the final directory brief.
- Never output your analysis, reasoning, plans, or conversational text.
- Write 35–65 words in plain text.
- Explain the directory's purpose and organization using the supplied context.
- Treat every subdirectory name as a separate directory.
- Do not invent files, folders, file contents, or filesystem changes.
- Do not recommend actions or discuss where a file should be moved.
- If evidence is limited, describe only what the evidence supports.
- Do not include a heading, bullet points, or Markdown.
`

	req := providers.LlmRequest{
		Provider: provider,
		Model:    provider.ActiveModel,
		Stream:   false,
		Messages: []providers.LlmMessage{
			{Role: "system", Content: query},
			{Role: "user", Content: feed},
		},
	}
	sctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	resp, err := req.SendTo(sctx, req.Model)
	if err != nil {
		return err
	}
	respContent, err := helpers.ExtractContent(*resp)
	if err != nil {
		return err
	}

	d.Mu.RLock()
	sum := sha256.Sum256([]byte(d.Path + time.Now().String()))
	id := hex.EncodeToString(sum[:])

	var inf = DirInfo{Path: d.Path,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Brief:     respContent,
		Modifier:  &InfoProvider{Provider: provider.Name, Model: req.Model}, DirId: id}

	loggy.AddNewf("Info generated! for %v <%v...>", loggy.PdLInfo, d.Path, respContent)
	err = d.WriteNewHelperFile(d.Info)
	if err != nil {
		return err
	}
	d.Mu.RUnlock()
	d.SetInfo(&inf)
	return nil
}

// func (d *Directory) UpdateHelperFile() error {
// 	b, err := yaml.Marshal(d.Info)
// 	if err != nil {
// 		return err
// 	}
// 	header := []byte("# This file is Managed by Clonopher. Manual edits may be overwritten.\n\n")
// 	content := append(header, b...)
// 	err = os.WriteFile(filepath.Join(d.Path, internal.HelperFileName), content, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (d *Directory) WriteNewHelperFile(inf *DirInfo) error {
	b, err := yaml.Marshal(inf)
	if err != nil {
		return err
	}
	header := []byte("# This file is Managed by Clonopher. Manual edits may be overwritten.\n\n")
	content := append(header, b...)
	err = d.Submit()
	if err != nil {
		return err
	}
	path := filepath.Join(d.Path, internal.HelperFileName)
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *DirInfo) Render() string {
	if s == nil {
		return ""
	}
	infoOptions := helpers.InfoOptions{
		&helpers.InfoOption{Info: `brief description: %v`, Options: []any{s.Brief}},
	}
	return infoOptions.ConcatIfExists()
}
