package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinburke/ssh_config"
	"github.com/synoxisllc/ossh/internal/tui"
)

func TestHome_EmptyConfig(t *testing.T) {
	h := NewHome(nil, nil, t.TempDir())
	view := h.View()

	if !strings.Contains(view, "No hosts configured yet") {
		t.Errorf("expected 'No hosts' message in view, got:\n%s", view)
	}
}

func TestHome_WithHosts_NoQuery(t *testing.T) {
	cfg := makeTestConfig(t, `
Host work-vpn
    HostName vpn.company.com
    User gana

Host personal
    HostName home.example.com
    User user
`)

	h := NewHome(cfg, nil, t.TempDir())
	view := h.View()

	if !strings.Contains(view, "work-vpn") {
		t.Errorf("expected 'work-vpn' in view")
	}
	if !strings.Contains(view, "personal") {
		t.Errorf("expected 'personal' in view")
	}
}

func TestHome_FuzzyFilter(t *testing.T) {
	cfg := makeTestConfig(t, `
Host work-vpn
    HostName vpn.company.com
    User gana

Host personal
    HostName home.example.com
    User user
`)

	h := NewHome(cfg, nil, t.TempDir())

	// Type "work"
	for _, r := range "work" {
		screen, _ := h.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		h = screen.(*Home)
	}

	view := h.View()
	if !strings.Contains(view, "work-vpn") {
		t.Errorf("expected 'work-vpn' in filtered results")
	}
	if strings.Contains(view, "personal") {
		t.Errorf("did not expect 'personal' in filtered results")
	}
}

func TestHome_CursorNavigation(t *testing.T) {
	cfg := makeTestConfig(t, `
Host host1
Host host2
Host host3
`)

	h := NewHome(cfg, nil, t.TempDir())

	// Initial cursor is 0
	if h.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", h.cursor)
	}

	// Move down
	screen, _ := h.Update(tea.KeyMsg{Type: tea.KeyDown})
	h = screen.(*Home)
	if h.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", h.cursor)
	}

	// Move down again
	screen, _ = h.Update(tea.KeyMsg{Type: tea.KeyDown})
	h = screen.(*Home)
	if h.cursor != 2 {
		t.Errorf("expected cursor at 2 after down, got %d", h.cursor)
	}

	// Try to move down beyond end (should clamp)
	screen, _ = h.Update(tea.KeyMsg{Type: tea.KeyDown})
	h = screen.(*Home)
	if h.cursor != 2 {
		t.Errorf("expected cursor to stay at 2 (clamped), got %d", h.cursor)
	}

	// Move up
	screen, _ = h.Update(tea.KeyMsg{Type: tea.KeyUp})
	h = screen.(*Home)
	if h.cursor != 1 {
		t.Errorf("expected cursor at 1 after up, got %d", h.cursor)
	}

	// Move up again
	screen, _ = h.Update(tea.KeyMsg{Type: tea.KeyUp})
	h = screen.(*Home)
	if h.cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", h.cursor)
	}

	// Try to move up beyond start (should clamp)
	screen, _ = h.Update(tea.KeyMsg{Type: tea.KeyUp})
	h = screen.(*Home)
	if h.cursor != 0 {
		t.Errorf("expected cursor to stay at 0 (clamped), got %d", h.cursor)
	}
}

func TestHome_EnterConnects(t *testing.T) {
	cfg := makeTestConfig(t, `
Host target
    HostName example.com
`)

	h := NewHome(cfg, nil, t.TempDir())

	// Press enter
	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from enter key")
	}

	// Execute the cmd to get the message
	msg := cmd()
	connectMsg, ok := msg.(tui.ConnectMsg)
	if !ok {
		t.Fatalf("expected ConnectMsg, got %T", msg)
	}

	if connectMsg.HostAlias != "target" {
		t.Errorf("expected HostAlias 'target', got %q", connectMsg.HostAlias)
	}
}

func TestHome_EscClearsQuery(t *testing.T) {
	cfg := makeTestConfig(t, `
Host host1
Host host2
`)

	h := NewHome(cfg, nil, t.TempDir())

	// Type a query
	screen, _ := h.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	h = screen.(*Home)
	if h.query != "h" {
		t.Errorf("expected query 'h', got %q", h.query)
	}

	// Press Esc (should clear query)
	screen, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	h = screen.(*Home)
	if cmd != nil {
		t.Error("expected no quit cmd when clearing query")
	}
	if h.query != "" {
		t.Errorf("expected empty query after esc, got %q", h.query)
	}
	if h.matches != nil {
		t.Error("expected matches to be nil after clearing query")
	}
}

func TestHome_EscQuitsWhenQueryEmpty(t *testing.T) {
	cfg := makeTestConfig(t, `
Host host1
`)

	h := NewHome(cfg, nil, t.TempDir())

	// Press Esc with empty query (should quit)
	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("expected tea.Quit cmd when esc pressed with empty query")
	}
}

// Helper to create a test SSH config from a string
func makeTestConfig(t *testing.T, configText string) *ssh_config.Config {
	t.Helper()
	cfg, err := ssh_config.Decode(strings.NewReader(configText))
	if err != nil {
		t.Fatalf("failed to parse test config: %v", err)
	}
	return cfg
}
