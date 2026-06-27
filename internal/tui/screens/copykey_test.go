package screens

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/synoxisllc/ossh/internal/keys"
	"github.com/synoxisllc/ossh/internal/tui"
)

func testHost() hostItem {
	return hostItem{Alias: "homelab", Hostname: "192.168.1.1", User: "gana"}
}

func setupSSHDirWithKey(t *testing.T) string {
	t.Helper()
	sshDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sshDir, "mykey"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "mykey.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}
	return sshDir
}

func TestCopyKey_InitRendersKeySelect(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())
	view := g.View()
	if !strings.Contains(view, "Copy key to homelab") {
		t.Errorf("expected title in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Select key:") {
		t.Errorf("expected key select prompt in view, got:\n%s", view)
	}
}

func TestCopyKey_EmptyKeysShowsEmptyState(t *testing.T) {
	g := NewCopyKey(t.TempDir(), nil, nil, testHost())
	view := g.View()
	if !strings.Contains(view, "No keys found") {
		t.Errorf("expected empty state in view, got:\n%s", view)
	}
}

func TestCopyKey_EmptyKeysEnterGoesHome(t *testing.T) {
	g := NewCopyKey(t.TempDir(), nil, nil, testHost())

	_, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected SwitchScreenMsg cmd")
	}
	msg := cmd()
	switchMsg, ok := msg.(tui.SwitchScreenMsg)
	if !ok {
		t.Fatalf("expected SwitchScreenMsg, got %T", msg)
	}
	if _, ok := switchMsg.Next.(*Home); !ok {
		t.Errorf("expected Home screen, got %T", switchMsg.Next)
	}
}

func TestCopyKey_KeySelectNavigates(t *testing.T) {
	sshDir := t.TempDir()
	for _, name := range []string{"keya", "keyb"} {
		if err := os.WriteFile(filepath.Join(sshDir, name), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	g := NewCopyKey(sshDir, nil, nil, testHost())
	if g.keyChoice != 0 {
		t.Fatalf("expected initial keyChoice 0, got %d", g.keyChoice)
	}

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyDown})
	g = screen.(*CopyKey)
	if g.keyChoice != 1 {
		t.Errorf("expected keyChoice 1 after down, got %d", g.keyChoice)
	}

	screen, _ = g.Update(tea.KeyMsg{Type: tea.KeyUp})
	g = screen.(*CopyKey)
	if g.keyChoice != 0 {
		t.Errorf("expected keyChoice 0 after up, got %d", g.keyChoice)
	}
}

func TestCopyKey_KeySelectEnterAdvancesToConfirm(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*CopyKey)
	if g.step != stepCKConfirm {
		t.Errorf("expected stepCKConfirm, got %d", g.step)
	}
	if g.fallback == "" {
		t.Error("expected fallback to be computed")
	}
}

func TestCopyKey_KeySelectEscGoesHome(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())

	_, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected SwitchScreenMsg cmd")
	}
	msg := cmd()
	if _, ok := msg.(tui.SwitchScreenMsg); !ok {
		t.Fatalf("expected SwitchScreenMsg, got %T", msg)
	}
}

func TestCopyKey_ConfirmEscGoesBackToKeySelect(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())
	g.step = stepCKConfirm

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEsc})
	g = screen.(*CopyKey)
	if g.step != stepCKKeySelect {
		t.Errorf("expected stepCKKeySelect, got %d", g.step)
	}
}

func TestCopyKey_ConfirmEnterReturnsCmd(t *testing.T) {
	if _, err := exec.LookPath("ssh-copy-id"); err != nil {
		t.Skip("ssh-copy-id not on PATH")
	}

	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())
	g.step = stepCKConfirm
	g.fallback = keys.ManualFallback(g.copyOpts())

	_, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected ExecProcess cmd from confirm enter")
	}
}

func TestCopyKey_FinishedWithErrGoesError(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())
	g.step = stepCKConfirm
	g.fallback = "fallback cmd"

	screen, _ := g.Update(copyKeyFinishedMsg{err: errors.New("copy failed")})
	g = screen.(*CopyKey)
	if g.step != stepCKError {
		t.Errorf("expected stepCKError, got %d", g.step)
	}
	if g.err == nil {
		t.Error("expected error to be set")
	}
}

func TestCopyKey_FinishedSuccessGoesDone(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())
	g.step = stepCKConfirm

	screen, _ := g.Update(copyKeyFinishedMsg{})
	g = screen.(*CopyKey)
	if g.step != stepCKDone {
		t.Errorf("expected stepCKDone, got %d", g.step)
	}
}

func TestCopyKey_DoneEnterGoesHome(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())
	g.step = stepCKDone

	_, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected SwitchScreenMsg cmd")
	}
	msg := cmd()
	switchMsg, ok := msg.(tui.SwitchScreenMsg)
	if !ok {
		t.Fatalf("expected SwitchScreenMsg, got %T", msg)
	}
	if _, ok := switchMsg.Next.(*Home); !ok {
		t.Errorf("expected Home screen, got %T", switchMsg.Next)
	}
}

func TestCopyKey_ErrorEnterGoesHome(t *testing.T) {
	sshDir := setupSSHDirWithKey(t)
	g := NewCopyKey(sshDir, nil, nil, testHost())
	g.step = stepCKError
	g.err = errors.New("copy failed")

	_, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected SwitchScreenMsg cmd")
	}
	msg := cmd()
	switchMsg, ok := msg.(tui.SwitchScreenMsg)
	if !ok {
		t.Fatalf("expected SwitchScreenMsg, got %T", msg)
	}
	if _, ok := switchMsg.Next.(*Home); !ok {
		t.Errorf("expected Home screen, got %T", switchMsg.Next)
	}
}
