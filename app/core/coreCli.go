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
)

var (
	payloadTUITitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	payloadTIDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	payloadTISuccess     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	payloadTIError       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	payloadTIWarn        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	payloadTIState       = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	// ErrPayloadTUIInterrupted is returned by Wait when the user presses Ctrl+C
	// inside the TUI. The caller should cancel the payload-processing context.
	ErrPayloadTUIInterrupted = errors.New("payload TUI interrupted by user")
)

// PayloadTUI manages one terminal monitor for all PayloadStatus entries in Stats.
// Start it only in CLI mode; API mode does not create a terminal UI or goroutine.
type PayloadTUI struct {
	program     *tea.Program
	disabled    bool
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
}

type payloadTUIFrameMsg string
type payloadTUIRefreshMsg time.Time
type payloadTUIStopMsg struct{}

// StartPayloadTUI starts the TUI asynchronously in CLI mode, then returns as
// soon as Bubble Tea has initialized. This is important: Program.Run/Start are
// blocking and must not run on the caller's goroutine, or the CLI work will not
// continue until the UI exits.
//
// done should be closed when payload processing has finished. An optional
// onInterrupt callback can cancel that processing when the user presses Ctrl+C.
// For example: StartPayloadTUI(true, done, cancel), where cancel is a context
// cancellation function. Existing two-argument calls remain valid.
func StartPayloadTUI(cli bool, done <-chan struct{}, onInterrupt ...func()) (*PayloadTUI, error) {
	if !cli {
		return &PayloadTUI{disabled: true}, nil
	}
	if done == nil {
		return nil, errors.New("payload TUI requires a non-nil done channel")
	}

	t := &PayloadTUI{
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

	// Init is called after Bubble Tea initializes the terminal. Also handle a
	// startup failure, where Init never runs and ready is never closed.
	select {
	case <-ready:
		return t, nil
	case <-t.runDone:
		select {
		case <-ready:
			return t, nil
		default:
		}
		err := t.error()
		if err != nil {
			return nil, fmt.Errorf("start payload TUI: %w", err)
		}
		return nil, errors.New("payload TUI exited before initialization")
	}
}

// Wait waits until the TUI exits, then prints a compact summary. It returns
// ErrPayloadTUIInterrupted if Ctrl+C was pressed while the TUI was active.
func (t *PayloadTUI) Wait() error {
	if t == nil || t.disabled {
		return nil
	}
	if t.program == nil || t.runDone == nil {
		return errors.New("payload TUI was not started")
	}

	<-t.runDone
	printPayloadResults()

	select {
	case <-t.interrupted:
		return ErrPayloadTUIInterrupted
	default:
		return t.error()
	}
}

// Interrupted closes as soon as the user presses Ctrl+C in the TUI. CLI code
// can select on this channel to cancel its worker context.
func (t *PayloadTUI) Interrupted() <-chan struct{} {
	if t == nil || t.disabled || t.interrupted == nil {
		return nil
	}
	return t.interrupted
}

// Stop requests a graceful TUI shutdown. It does not cancel payload workers.
func (t *PayloadTUI) Stop() {
	if t == nil || t.disabled || t.program == nil {
		return
	}
	t.program.Quit()
}

func (t *PayloadTUI) error() error {
	t.errMu.RLock()
	defer t.errMu.RUnlock()
	return t.runErr
}

// spinner produces animated frames until done or stop receives/closes.
// Its buffer is intentionally small so obsolete frames are dropped.
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
					// Drop stale frames; keep the newest animation responsive.
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
	header := payloadTUITitleStyle.Render("Clonopher") + payloadTIDimStyle.Render("  •  live payload monitor")
	if len(m.rows) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			payloadTIDimStyle.Render(m.frame+" Waiting for payloads..."),
		)
	}

	width := m.width
	if width <= 0 {
		width = 100
	}
	lines := make([]string, 0, len(m.rows)+3)
	lines = append(lines, header, payloadTIDimStyle.Render(strings.Repeat("─", maxInt(20, minInt(width, 120)))))

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
		available := width - lipgloss.Width(prefix) - lipgloss.Width(suffix) - 2
		path = middleTruncate(path, available)
		lines = append(lines, prefix+path+suffix)
	}

	lines = append(lines, "", payloadTIDimStyle.Render("Ctrl+C cancels the CLI only when an interrupt callback is provided."))
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
		}
		if status.Stage != nil {
			// Stage.Path is expected to be immutable after Stage construction.
			row.stagePath = status.Stage.Path
		}
		status.Mu.RUnlock()
		rows = append(rows, row)
	}

	return rows
}

func printPayloadResults() {
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
		fmt.Printf("%s %-12s state=%s action=%s path=%s\n", marker, row.name, row.state, row.lastAction, finalPath)
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
