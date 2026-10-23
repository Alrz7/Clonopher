package core

import (
	"clonopher/app/burrow"
	"clonopher/app/providers"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func (i *ObjectInfo) Render() string {
	if i == nil {
		return ""
	}
	b, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return string(b)
}

type NewStageProposal struct {
	Name  string `json:"name"`
	Brief string `json:"brief"`
}

func parseProposal(content string) (*NewStageProposal, error) {
	var p NewStageProposal

	if err := json.Unmarshal([]byte(content), &p); err != nil {
		return nil, fmt.Errorf("decode directory proposal: %w", err)
	}

	p.Name = strings.TrimSpace(p.Name)
	p.Brief = strings.TrimSpace(p.Brief)

	if p.Name == "" || p.Brief == "" {
		return nil, errors.New("incomplete directory proposal")
	}

	// The name must be a single directory name, not a path.
	if p.Name == "." || p.Name == ".." ||
		filepath.Base(p.Name) != p.Name ||
		strings.ContainsAny(p.Name, `/\\`) {
		return nil, errors.New("invalid directory name")
	}

	return &p, nil
}

func (p *Payload) NewStageWithInfo() error {
	// path := p.Stage.Path
	// p.Stage.InfMu.Lock()
	// defer p.Stage.InfMu.Unlock()
	// if p.Stage.Path != path {
	// 	return nil
	// } each call has its own parameter like diferent file infos and data; so each call creates a new subStage on the parrent stage
	// therefor we may not need a mutex for this.

	loggy.Info("generating Info...")
	feed := helpers.InfoOptions{
		&helpers.InfoOption{Info: `file-name: %v`, Options: []any{p.Name}},
		&helpers.InfoOption{Info: `file-info: %v`, Options: []any{p.Info.Render()}},
		&helpers.InfoOption{Info: `parentDir-path: %v`, Options: []any{p.Stage.Path}},
		&helpers.InfoOption{Info: `sub-dirs: %v`, Options: []any{slices.Collect(maps.Keys(p.Stage.Roots))}},
		&helpers.InfoOption{Info: `dir-files: %v`, Options: []any{p.Stage.Files}},
	}.ConcatIfExists()

	query := `You are Clonopher's directory metadata composer.

A new directory is being created to organize a target file.
Generate a useful, reusable brief describing the directory's intended
purpose & a meaningfull name for new directory. This brief will help classify files into this directory in
the future.

Evidence priority:
1. file-info is the primary source for understanding the target file's
   subject and purpose. Preserve its meaning and use it as the main
   basis for defining the directory's purpose.
2. Use file-name as a secondary clue, especially when file-info is absent
   or incomplete.
3. Use dir-path, sub-dirs, and dir-files only as additional context.
   Names alone do not prove what files contain or what directories are for.

Writing requirements:
- Write 40–70 words in 2–4 complete sentences.
- Describe the directory's intended purpose, subject, and the kinds of
  related resources that would reasonably belong there.
- Make the description broad enough to accommodate related future files,
  but specific enough to distinguish this directory from unrelated topics.
- Base the description primarily on the target file's information.
  Do not merely restate its filename or describe only one file.
- Use existing directory context to avoid proposing a purpose that
  duplicates an obviously better-established sibling category.
- Treat every directory name as a separate entry. Never combine multiple
  names into a single path or directory.
- Never invent file contents, existing resources, technologies, or
  intended uses that are not reasonably supported by the input.
- The directory is being created now. Do not claim it already contains
  files or resources unless the supplied context explicitly establishes that.
- Do not recommend actions, choose destinations, or explain your reasoning.

Naming requirements:
- Return a concise, meaningful directory name.
- Follow the naming conventions of existing sibling directories.
- Do not include a filesystem path, slashes, or a file extension.
- Avoid names that duplicate existing sibling directories.
- Name the directory after its subject, not after a single filename.
- Use the target file's description as the primary source of evidence.

Output exactly one valid JSON object with these fields:
{
  "name": "directory name",
  "brief": "directory description"
}
Return JSON only. Do not include Markdown, headings, or explanations.`

	req := providers.LlmRequest{
		Provider: p.LlmProvider,
		Model:    p.LlmProvider.ActiveModel,
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
	content, err := helpers.ExtractContent(*resp)
	if err != nil {
		return err
	}

	respContent, err := parseProposal(content)
	if err != nil {
		return err
	}
	/*
		we need to check for dupllications, ai bugs, ... every commot errors that can happen, Not only here...
	*/
	newPath := filepath.Join(p.Stage.Path, respContent.Name)
	newStage := burrow.New(newPath)

	newStage.SetInfo(&burrow.StageInfo{Path: newStage.Path, CreatedAt: time.Now(), UpdatedAt: time.Now(), Brief: respContent.Brief, Creator: &burrow.InfoProvider{Provider: p.LlmProvider.Name, Model: req.Model}})
	loggy.Infof("Info generated! for %v <%v...>", p.Name, respContent)
	err = newStage.WriteNewHelperFile(newStage.Info)
	if err != nil {
		return err
	}
	p.Stage.AddRoot(respContent.Name, newStage)

	p.SetNext(newStage)
	return nil
}
