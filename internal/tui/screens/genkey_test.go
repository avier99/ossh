package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/synoxisllc/ossh/internal/tui"
)

func TestGenKey_InitRendersNameStep(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	view := g.View()
	if !strings.Contains(view, "New SSH key") {
		t.Errorf("expected name step title in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Key name:") {
		t.Errorf("expected name prompt in view, got:\n%s", view)
	}
}

func TestGenKey_NameEnterAdvancesToFolder(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.nameInput.SetValue("mykey")

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*GenKey)
	if g.step != stepFolder {
		t.Errorf("expected stepFolder, got %d", g.step)
	}
	if g.name != "mykey" {
		t.Errorf("expected name 'mykey', got %q", g.name)
	}
}

func TestGenKey_FolderBlankSkipsToType(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepFolder
	g.name = "mykey"

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*GenKey)
	if g.step != stepType {
		t.Errorf("expected stepType, got %d", g.step)
	}
	if g.subDir != "" {
		t.Errorf("expected empty subDir, got %q", g.subDir)
	}
}

func TestGenKey_FolderValueAdvancesToType(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepFolder
	g.name = "mykey"
	g.folderInput.SetValue("homelab-keys")

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*GenKey)
	if g.step != stepType {
		t.Errorf("expected stepType, got %d", g.step)
	}
	if g.subDir != "homelab-keys" {
		t.Errorf("expected subDir 'homelab-keys', got %q", g.subDir)
	}
}

func TestGenKey_FolderEscGoesBackToName(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepFolder

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEsc})
	g = screen.(*GenKey)
	if g.step != stepName {
		t.Errorf("expected stepName, got %d", g.step)
	}
}

func TestGenKey_FolderInvalidBlocked(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepFolder
	g.name = "mykey"
	g.folderInput.SetValue("foo/bar")

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*GenKey)
	if g.step != stepFolder {
		t.Errorf("expected to stay on stepFolder, got %d", g.step)
	}
	if g.folderErr == nil {
		t.Error("expected folder validation error")
	}
}

func TestGenKey_EmptyNameBlocked(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*GenKey)
	if g.step != stepName {
		t.Errorf("expected to stay on stepName, got %d", g.step)
	}
	if g.nameErr == nil {
		t.Error("expected name validation error")
	}
}

func TestGenKey_EscFromNameCancels(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)

	_, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEsc})
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

func TestGenKey_TypeSelectionNavigates(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepType
	g.typeChoice = 0

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyDown})
	g = screen.(*GenKey)
	if g.typeChoice != 1 {
		t.Errorf("expected typeChoice 1 after down, got %d", g.typeChoice)
	}

	screen, _ = g.Update(tea.KeyMsg{Type: tea.KeyUp})
	g = screen.(*GenKey)
	if g.typeChoice != 0 {
		t.Errorf("expected typeChoice 0 after up, got %d", g.typeChoice)
	}
}

func TestGenKey_TypeEnterAdvances(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepType
	g.name = "mykey"
	g.typeChoice = 0

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*GenKey)
	if g.step != stepConfirm {
		t.Errorf("expected stepConfirm, got %d", g.step)
	}
}

func TestGenKey_TypeEscGoesBack(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepType

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEsc})
	g = screen.(*GenKey)
	if g.step != stepFolder {
		t.Errorf("expected stepFolder, got %d", g.step)
	}
}

func TestGenKey_ConfirmEnterReturnsCmd(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepConfirm
	g.name = "mykey"
	g.keyType = "ed25519"

	_, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected ExecProcess cmd from confirm enter")
	}
}

func TestGenKey_ConfirmEscGoesBack(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepConfirm

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEsc})
	g = screen.(*GenKey)
	if g.step != stepType {
		t.Errorf("expected stepType, got %d", g.step)
	}
}

func TestGenKey_WarnNoPassphrase_EnterAdvances(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepWarnNoPassphrase
	g.name = "mykey"

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*GenKey)
	if g.step != stepDone {
		t.Errorf("expected stepDone, got %d", g.step)
	}
}

func TestGenKey_DoneEnterReturnsHome(t *testing.T) {
	g := NewGenKey(t.TempDir(), nil, nil)
	g.step = stepDone

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
