package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	payloadTUIRefreshRate = 120 * time.Millisecond
	payloadTUISpinnerRate = 80 * time.Millisecond
	payloadTUILogVisible  = 8
)

var (
	payloadTUITitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	payloadTIDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	payloadTISuccess     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#16b48f"))
	payloadTIError       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#973245"))
	payloadTIWarn        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b59232"))
	payloadTIState       = lipgloss.NewStyle().Foreground(lipgloss.Color("#52c1c5"))

	ErrPayloadTUIInterrupted = errors.New("payload TUI interrupted by user")
)

// PayloadTUI monitors all PayloadStatus entries in Stats.
type PayloadTUI struct {
	program     *tea.Program
	logLevel    string
	runDone     chan struct{}
	stopFrames  chan struct{}
	interrupted chan struct{}
	interrupt   sync.Once
	errMu       sync.RWMutex
	runErr      error
}

type payloadTUIModel struct {
	frames          <-chan string
	frame           string
	rows            []payloadTUIRow
	width           int
	logLevel        string
	ready           chan struct{}
	signalInterrupt func()
}

type payloadTUIRow struct {
	id         string
	name       string
	state      PayloadState
	stagePath  string
	lastAction string
	iteration  int
	logs       []PayloadLog
}

type payloadTUILogRow struct {
	payloadName string
	log         PayloadLog
}

type payloadTUIFrameMsg string
type payloadTUIRefreshMsg time.Time
type payloadTUIStopMsg struct{}

// StartPayloadTUI starts the TUI asynchronously.
//
// done is closed when payload processing has finished.
// logLevel controls the live log console:
//   - "" or an unknown value: no live log console
//   - "error": only errors
//   - "warn": only warnings
//   - "info": only info logs
//   - "all" or "silly": all log types
//
// Final output always includes errors. "warn" also includes warnings,
// "info" also includes info logs, and "all"/"silly" include every log type.
//
// An optional onInterrupt callback can cancel payload processing when Ctrl+C
// is pressed inside the TUI.
func StartPayloadTUI(
	done <-chan struct{},
	logLevel string,
	onInterrupt ...func(),
) (*PayloadTUI, error) {
	if done == nil {
		return nil, errors.New("payload TUI requires a non-nil done channel")
	}

	t := &PayloadTUI{
		logLevel:    normalizePayloadLogLevel(logLevel),
		runDone:     make(chan struct{}),
		stopFrames:  make(chan struct{}),
		interrupted: make(chan struct{}),
	}

	var interruptCallback func()
	if len(onInterrupt) > 0 {
		interruptCallback = onInterrupt[0]
	}

	signalInterrupt := func() {
		t.interrupt.Do(func() {
			close(t.interrupted)

			if interruptCallback != nil {
				interruptCallback()
			}
		})
	}

	ready := make(chan struct{})

	model := payloadTUIModel{
		frames:          spinner(done, t.stopFrames),
		frame:           "⠋",
		width:           100,
		logLevel:        t.logLevel,
		ready:           ready,
		signalInterrupt: signalInterrupt,
	}

	t.program = tea.NewProgram(model, tea.WithAltScreen())

	go func() {
		_, err := t.program.Run()

		if errors.Is(err, tea.ErrInterrupted) {
			signalInterrupt()
		}

		t.errMu.Lock()
		t.runErr = err
		t.errMu.Unlock()

		close(t.stopFrames)
		close(t.runDone)
	}()

	select {
	case <-ready:
		return t, nil

	case <-t.runDone:
		select {
		case <-ready:
			return t, nil
		default:
		}

		if err := t.error(); err != nil {
			return nil, fmt.Errorf("start payload TUI: %w", err)
		}

		return nil, errors.New("payload TUI exited before initialization")
	}
}

// Wait waits for the TUI to exit and prints payload results and filtered logs.
func (t *PayloadTUI) Wait() error {
	if t == nil {
		return nil
	}

	if t.program == nil || t.runDone == nil {
		return errors.New("payload TUI was not started")
	}

	<-t.runDone

	printPayloadResults(t.logLevel)

	select {
	case <-t.interrupted:
		return ErrPayloadTUIInterrupted
	default:
		return t.error()
	}
}

// Interrupted closes when Ctrl+C is pressed inside the TUI.
func (t *PayloadTUI) Interrupted() <-chan struct{} {
	if t == nil || t.interrupted == nil {
		return nil
	}

	return t.interrupted
}

// Stop requests a graceful TUI shutdown. It does not cancel payload workers.
func (t *PayloadTUI) Stop() {
	if t == nil || t.program == nil {
		return
	}

	t.program.Quit()
}

func (t *PayloadTUI) error() error {
	t.errMu.RLock()
	defer t.errMu.RUnlock()

	return t.runErr
}

func normalizePayloadLogLevel(level string) string {
	level = strings.ToLower(strings.TrimSpace(level))

	switch level {
	case "error", "warn", "info", "all", "silly":
		return level
	default:
		return ""
	}
}

func hasLivePayloadLogConsole(level string) bool {
	switch normalizePayloadLogLevel(level) {
	case "error", "warn", "info", "all", "silly":
		return true
	default:
		return false
	}
}

// Live console: error, warn and info each display only their own type.
// all and silly display every log type.
func livePayloadLogVisible(level string, entry PayloadLog) bool {
	switch normalizePayloadLogLevel(level) {
	case "error":
		return entry.Type == PdLError
	case "warn":
		return entry.Type == PdLWarn
	case "info":
		return entry.Type == PdLInfo
	case "all", "silly":
		return true
	default:
		return false
	}
}

// Final output always includes errors. warn adds warnings, info adds info logs,
// and all/silly include every log type.
func finalPayloadLogVisible(level string, entry PayloadLog) bool {
	switch normalizePayloadLogLevel(level) {
	case "warn":
		return entry.Type == PdLError || entry.Type == PdLWarn
	case "info":
		return entry.Type == PdLError || entry.Type == PdLInfo
	case "all", "silly":
		return true
	default:
		return entry.Type == PdLError
	}
}

func spinner(done <-chan struct{}, stop <-chan struct{}) <-chan string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	out := make(chan string, 1)

	go func() {
		defer close(out)

		ticker := time.NewTicker(payloadTUISpinnerRate)
		defer ticker.Stop()

		index := 0

		for {
			select {
			case <-done:
				return
			case <-stop:
				return
			case <-ticker.C:
				frame := frames[index]
				index = (index + 1) % len(frames)

				select {
				case out <- frame:
				case <-done:
					return
				case <-stop:
					return
				default:
					// Drop stale frames to keep animation responsive.
				}
			}
		}
	}()

	return out
}

func (m payloadTUIModel) Init() tea.Cmd {
	if m.ready != nil {
		close(m.ready)
	}

	return tea.Batch(m.waitForFrame(), m.refresh())
}

func (m payloadTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case payloadTUIFrameMsg:
		m.frame = string(msg)
		return m, m.waitForFrame()

	case payloadTUIRefreshMsg:
		m.rows = readPayloadTUIRows()
		return m, m.refresh()

	case payloadTUIStopMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.signalInterrupt != nil {
				m.signalInterrupt()
			}

			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	}

	return m, nil
}

func (m payloadTUIModel) View() string {
	header := payloadTUITitleStyle.Render("Clonopher") +
		payloadTIDimStyle.Render("  •  live payload monitor")

	width := m.width
	if width <= 0 {
		width = 100
	}

	separator := payloadTIDimStyle.Render(
		strings.Repeat("─", maxInt(20, minInt(width, 120))),
	)

	lines := []string{header, separator}

	if len(m.rows) == 0 {
		lines = append(lines, payloadTIDimStyle.Render(
			m.frame+" Waiting for payloads...",
		))
	} else {
		for _, row := range m.rows {
			indicator := m.frame

			switch row.state {
			case PayloadStateCompleted:
				indicator = payloadTISuccess.Render("✓")
			case PayloadStateFailed:
				indicator = payloadTIError.Render("✗")
			case PayloadStateCancelled:
				indicator = payloadTIDimStyle.Render("⊘")
			case PayloadStateUncertain:
				indicator = payloadTIWarn.Render("?")
			}

			state := "[" + string(row.state) + "]"
			if row.state == "" {
				state = "[initializing]"
			}
			state = payloadTIState.Render(state)

			path := row.stagePath
			if path == "" {
				path = "—"
			}

			action := row.lastAction
			if action == "" {
				action = "—"
			}

			prefix := indicator + " " + state + " "
			suffix := " " + action
			available := width -
				lipgloss.Width(prefix) -
				lipgloss.Width(suffix) - 2

			path = middleTruncate(path, available)
			lines = append(lines, prefix+path+suffix)
		}
	}

	if hasLivePayloadLogConsole(m.logLevel) {
		lines = append(
			lines,
			"",
			payloadTUITitleStyle.Render("Payload logs"),
			separator,
		)

		logs := readPayloadTUILogs(m.rows, m.logLevel, payloadTUILogVisible)

		if len(logs) == 0 {
			lines = append(lines, payloadTIDimStyle.Render(
				"No matching payload logs yet.",
			))
		} else {
			for _, entry := range logs {
				lines = append(lines, renderPayloadLogLine(entry, width))
			}
		}
	}

	lines = append(
		lines,
		"",
		payloadTIDimStyle.Render(
			"Ctrl+C cancels the CLI only when an interrupt callback is provided.",
		),
	)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m payloadTUIModel) waitForFrame() tea.Cmd {
	return func() tea.Msg {
		frame, ok := <-m.frames
		if !ok {
			return payloadTUIStopMsg{}
		}

		return payloadTUIFrameMsg(frame)
	}
}

func (m payloadTUIModel) refresh() tea.Cmd {
	return tea.Tick(payloadTUIRefreshRate, func(now time.Time) tea.Msg {
		return payloadTUIRefreshMsg(now)
	})
}

// Copy logs while holding the status lock so rendering can't race with appends.
func readPayloadTUIRows() []payloadTUIRow {
	statuses := Stats.Snapshot()
	rows := make([]payloadTUIRow, 0, len(statuses))

	for _, status := range statuses {
		if status == nil || status.Mu == nil {
			continue
		}

		status.Mu.RLock()

		row := payloadTUIRow{
			id:         status.ID,
			name:       status.Name,
			state:      status.State,
			lastAction: status.LastAction,
			iteration:  status.Iteration,
			logs:       append([]PayloadLog(nil), status.Logs...),
		}

		if status.Stage != nil {
			// Stage.Path is expected to be immutable after construction.
			row.stagePath = status.Stage.Path
		}

		status.Mu.RUnlock()
		rows = append(rows, row)
	}

	return rows
}

func readPayloadTUILogs(
	rows []payloadTUIRow,
	level string,
	limit int,
) []payloadTUILogRow {
	logs := make([]payloadTUILogRow, 0)

	for _, row := range rows {
		for _, entry := range row.logs {
			if livePayloadLogVisible(level, entry) {
				logs = append(logs, payloadTUILogRow{
					payloadName: row.name,
					log:         entry,
				})
			}
		}
	}

	if limit > 0 && len(logs) > limit {
		logs = logs[len(logs)-limit:]
	}

	return logs
}

// Keep the type label and style driven by PayloadLog.Type, so new log types
// still display their actual type instead of needing a hard-coded label.
func renderPayloadLogLine(entry payloadTUILogRow, width int) string {
	label := strings.ToUpper(string(entry.log.Type))

	var styledType string

	switch entry.log.Type {
	case PdLError:
		styledType = payloadTIError.Render(label)
	case PdLWarn:
		styledType = payloadTIWarn.Render(label)
	case PdLInfo:
		styledType = payloadTIState.Render(label)
	default:
		styledType = payloadTIDimStyle.Render(label)
	}

	prefix := styledType + " [" + entry.payloadName + "] "
	content := strings.ReplaceAll(entry.log.Content, "\n", " ")
	available := width - lipgloss.Width(prefix)
	content = middleTruncate(content, available)

	return prefix + content
}

func printPayloadResults(logLevel string) {
	rows := readPayloadTUIRows()
	if len(rows) == 0 {
		return
	}

	fmt.Println("\nClonopher payload results:")

	for _, row := range rows {
		marker := "•"

		switch row.state {
		case PayloadStateCompleted:
			marker = payloadTISuccess.Render("✓")
		case PayloadStateFailed:
			marker = payloadTIError.Render("✗")
		case PayloadStateCancelled:
			marker = "⊘"
		case PayloadStateUncertain:
			marker = payloadTIWarn.Render("?")
		}

		finalPath := row.stagePath
		if row.name != "" && finalPath != "" {
			finalPath = filepath.Join(finalPath, row.name)
		}

		fmt.Printf(
			"%s %-12s state=%s action=%s path=%s\n",
			marker,
			row.name,
			row.state,
			row.lastAction,
			finalPath,
		)
	}

	var matched []payloadTUILogRow

	for _, row := range rows {
		for _, entry := range row.logs {
			if finalPayloadLogVisible(logLevel, entry) {
				matched = append(matched, payloadTUILogRow{
					payloadName: row.name,
					log:         entry,
				})
			}
		}
	}

	if len(matched) == 0 {
		return
	}

	fmt.Println("\nPayload logs:")

	for _, entry := range matched {
		fmt.Println(renderPayloadLogLine(entry, 120))
	}
}

func middleTruncate(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= maxWidth {
		return value
	}

	if maxWidth == 1 {
		return "…"
	}

	left := (maxWidth - 1) / 2
	right := maxWidth - left - 1

	return string(runes[:left]) + "…" + string(runes[len(runes)-right:])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
