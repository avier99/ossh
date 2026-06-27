package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAddHost_InitRendersAliasStep(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	view := g.View()
	if !strings.Contains(view, "Add host") {
		t.Errorf("expected title in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Alias:") {
		t.Errorf("expected alias prompt in view, got:\n%s", view)
	}
}

func TestAddHost_EmptyAliasBlocked(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHAlias {
		t.Errorf("expected to stay on stepAHAlias, got %d", g.step)
	}
	if g.aliasErr == nil {
		t.Error("expected alias validation error")
	}
}

func TestAddHost_AliasEnterAdvancesToHostname(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.aliasInput.SetValue("homelab")

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHHostname {
		t.Errorf("expected stepAHHostname, got %d", g.step)
	}
	if g.alias != "homelab" {
		t.Errorf("expected alias 'homelab', got %q", g.alias)
	}
}

func TestAddHost_HostnameEnterAdvancesToUser(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHHostname
	g.alias = "homelab"
	g.hostnameInput.SetValue("192.168.1.100")

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHUser {
		t.Errorf("expected stepAHUser, got %d", g.step)
	}
	if g.hostname != "192.168.1.100" {
		t.Errorf("expected hostname, got %q", g.hostname)
	}
}

func TestAddHost_UserBlankAdvancesToPort(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHUser
	g.alias = "homelab"
	g.hostname = "192.168.1.100"

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHPort {
		t.Errorf("expected stepAHPort, got %d", g.step)
	}
}

func TestAddHost_PortInvalidBlocked(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHPort
	g.portInput.SetValue("99999")

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHPort {
		t.Errorf("expected to stay on stepAHPort, got %d", g.step)
	}
	if g.portErr == nil {
		t.Error("expected port validation error")
	}
}

func TestAddHost_PortBlankAdvancesToAuthMethod(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHPort

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHAuthMethod {
		t.Errorf("expected stepAHAuthMethod, got %d", g.step)
	}
	if g.port != 0 {
		t.Errorf("expected port 0, got %d", g.port)
	}
}

func TestAddHost_AuthKeyAdvancesToKeyFile(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHAuthMethod
	g.authChoice = 0

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHKeyFile {
		t.Errorf("expected stepAHKeyFile, got %d", g.step)
	}
	if g.authMethod != "key" {
		t.Errorf("expected authMethod 'key', got %q", g.authMethod)
	}
}

func TestAddHost_AuthPasswordAdvancesToWarn(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHAuthMethod
	g.authChoice = 1

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHPasswordWarn {
		t.Errorf("expected stepAHPasswordWarn, got %d", g.step)
	}
	if g.authMethod != "password" {
		t.Errorf("expected authMethod 'password', got %q", g.authMethod)
	}
}

func TestAddHost_KeyFileSkipAdvancesToConfirm(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHKeyFile
	g.authMethod = "key"
	g.keyChoice = 0

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHConfirm {
		t.Errorf("expected stepAHConfirm, got %d", g.step)
	}
	if g.identityFile != "" {
		t.Errorf("expected empty identityFile, got %q", g.identityFile)
	}
}

func TestAddHost_PasswordWarnEnterAdvancesToConfirm(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHPasswordWarn
	g.authMethod = "password"

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = screen.(*AddHost)
	if g.step != stepAHConfirm {
		t.Errorf("expected stepAHConfirm, got %d", g.step)
	}
}

func TestAddHost_EscFromHostnameGoesBack(t *testing.T) {
	g := NewAddHost(t.TempDir(), nil, nil)
	g.step = stepAHHostname

	screen, _ := g.Update(tea.KeyMsg{Type: tea.KeyEsc})
	g = screen.(*AddHost)
	if g.step != stepAHAlias {
		t.Errorf("expected stepAHAlias, got %d", g.step)
	}
}
