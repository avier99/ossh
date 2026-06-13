package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinburke/ssh_config"
	"github.com/synoxisllc/ossh/internal/config"
	"github.com/synoxisllc/ossh/internal/keys"
	"github.com/synoxisllc/ossh/internal/tui"
)

type genKeyStep int

const (
	stepName genKeyStep = iota
	stepType
	stepConfirm
	stepWarnNoPassphrase
	stepDone
	stepError
)

// GenKey is the key generation wizard screen.
type GenKey struct {
	step       genKeyStep
	sshDir     string
	homeDir    string
	cfg        *ssh_config.Config
	findings   []config.Finding
	name       string
	keyType    keys.KeyType
	nameInput  textinput.Model
	typeChoice int
	nameErr    error
	err        error
}

// NewGenKey creates a new key generation wizard.
func NewGenKey(sshDir string, cfg *ssh_config.Config, findings []config.Finding) *GenKey {
	ti := textinput.New()
	ti.Placeholder = "mykey"
	ti.CharLimit = 64
	ti.Width = 30
	ti.Focus()

	homeDir, _ := os.UserHomeDir()

	return &GenKey{
		step:       stepName,
		sshDir:     sshDir,
		homeDir:    homeDir,
		cfg:        cfg,
		findings:   findings,
		typeChoice: 0,
		nameInput:  ti,
	}
}

// Init implements tui.Screen.
func (g *GenKey) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tui.Screen.
func (g *GenKey) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case genKeyFinishedMsg:
		return g.handleFinished(msg)

	case tea.KeyMsg:
		return g.handleKey(msg)
	}

	var cmd tea.Cmd
	if g.step == stepName {
		g.nameInput, cmd = g.nameInput.Update(msg)
	}
	return g, cmd
}

type genKeyFinishedMsg struct {
	err           error
	hasPassphrase bool
}

func (g *GenKey) handleFinished(msg genKeyFinishedMsg) (tui.Screen, tea.Cmd) {
	if msg.err != nil {
		g.step = stepError
		g.err = msg.err
		return g, nil
	}
	if !msg.hasPassphrase {
		g.step = stepWarnNoPassphrase
		return g, nil
	}
	g.step = stepDone
	return g, nil
}

func (g *GenKey) homeScreen() *Home {
	return NewHome(g.cfg, g.findings, g.sshDir)
}

func (g *GenKey) handleKey(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch g.step {
	case stepName:
		return g.handleKeyName(msg)
	case stepType:
		return g.handleKeyType(msg)
	case stepConfirm:
		return g.handleKeyConfirm(msg)
	case stepWarnNoPassphrase:
		return g.handleKeyWarn(msg)
	case stepDone:
		return g.handleKeyDone(msg)
	case stepError:
		return g.handleKeyError(msg)
	}
	return g, nil
}

func (g *GenKey) handleKeyName(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		return g, func() tea.Msg {
			return tui.SwitchScreenMsg{Next: g.homeScreen()}
		}
	case "enter":
		name := strings.TrimSpace(g.nameInput.Value())
		opts := keys.GenerateOpts{SSHDir: g.sshDir, Name: name}
		if err := keys.Validate(opts); err != nil {
			g.nameErr = err
			return g, nil
		}
		g.name = name
		g.nameErr = nil
		g.step = stepType
		return g, nil
	}

	var cmd tea.Cmd
	g.nameInput, cmd = g.nameInput.Update(msg)
	return g, cmd
}

func (g *GenKey) handleKeyType(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepName
		return g, nil
	case "up", "k":
		if g.typeChoice > 0 {
			g.typeChoice--
		}
		return g, nil
	case "down", "j":
		if g.typeChoice < 1 {
			g.typeChoice++
		}
		return g, nil
	case "enter":
		if g.typeChoice == 0 {
			g.keyType = keys.KeyTypeED25519
		} else {
			g.keyType = keys.KeyTypeRSA4096
		}
		g.step = stepConfirm
		return g, nil
	}
	return g, nil
}

func (g *GenKey) handleKeyConfirm(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepType
		return g, nil
	case "enter":
		opts := keys.GenerateOpts{
			SSHDir:  g.sshDir,
			Name:    g.name,
			KeyType: g.keyType,
		}
		cmd, err := keys.Command(opts)
		if err != nil {
			g.step = stepError
			g.err = err
			return g, nil
		}

		sshDir := g.sshDir
		name := g.name
		return g, tea.ExecProcess(cmd, func(execErr error) tea.Msg {
			if execErr != nil {
				return genKeyFinishedMsg{err: execErr}
			}
			if err := keys.Verify(sshDir, name); err != nil {
				return genKeyFinishedMsg{err: err}
			}
			hasPassphrase, err := keys.HasPassphrase(sshDir, name)
			if err != nil {
				return genKeyFinishedMsg{err: err}
			}
			return genKeyFinishedMsg{hasPassphrase: hasPassphrase}
		})
	}
	return g, nil
}

func (g *GenKey) handleKeyWarn(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		return g, func() tea.Msg {
			return tui.SwitchScreenMsg{Next: g.homeScreen()}
		}
	case "enter":
		g.step = stepDone
		return g, nil
	}
	return g, nil
}

func (g *GenKey) handleKeyDone(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc", "enter":
		return g, func() tea.Msg {
			return tui.SwitchScreenMsg{Next: g.homeScreen()}
		}
	}
	return g, nil
}

func (g *GenKey) handleKeyError(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc", "enter":
		return g, func() tea.Msg {
			return tui.SwitchScreenMsg{Next: g.homeScreen()}
		}
	}
	return g, nil
}

// View implements tui.Screen.
func (g *GenKey) View() string {
	switch g.step {
	case stepName:
		return g.viewName()
	case stepType:
		return g.viewType()
	case stepConfirm:
		return g.viewConfirm()
	case stepWarnNoPassphrase:
		return g.viewWarnNoPassphrase()
	case stepDone:
		return g.viewDone()
	case stepError:
		return g.viewError()
	}
	return ""
}

func (g *GenKey) viewName() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  New SSH key"))
	b.WriteString("\n\n")
	b.WriteString("  Key name: ")
	b.WriteString(g.nameInput.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("  Will create: %s", sshDisplayPath(g.homeDir, g.sshDir, g.nameInput.Value()))))
	b.WriteString("\n\n")
	if g.nameErr != nil {
		b.WriteString(warningStyle.Render("  " + g.nameErr.Error()))
		b.WriteString("\n\n")
	}
	b.WriteString(helpStyle.Render("  enter:next  esc:cancel"))
	return b.String()
}

func (g *GenKey) viewType() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Key type"))
	b.WriteString("\n\n")

	types := []struct {
		label string
		note  string
	}{
		{"ed25519", "(recommended)"},
		{"rsa 4096", "(legacy systems only)"},
	}
	for i, item := range types {
		prefix := "    "
		if i == g.typeChoice {
			prefix = "  ▶ "
		}
		line := fmt.Sprintf("%s%-12s  %s", prefix, item.label, item.note)
		if i == g.typeChoice {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter:select  ↑↓:choose  esc:back"))
	return b.String()
}

func (g *GenKey) viewConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Ready to generate"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Name:        %s\n", g.name))
	b.WriteString(fmt.Sprintf("  Type:        %s\n", keyTypeLabel(g.keyType)))
	b.WriteString(fmt.Sprintf("  Private key: %s       (0600)\n", sshDisplayPath(g.homeDir, g.sshDir, g.name)))
	b.WriteString(fmt.Sprintf("  Public key:  %s   (0644)\n", sshDisplayPath(g.homeDir, g.sshDir, g.name)+".pub"))
	b.WriteString("\n")
	b.WriteString("  ssh-keygen will prompt for a passphrase.\n\n")
	b.WriteString(helpStyle.Render("  enter:generate  esc:back"))
	return b.String()
}

func (g *GenKey) viewWarnNoPassphrase() string {
	var b strings.Builder
	b.WriteString(warningStyle.Render("  [!] No passphrase set"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Your private key has no passphrase. Anyone who can read\n  %s can use it without restriction.\n\n", sshDisplayPath(g.homeDir, g.sshDir, g.name)))
	b.WriteString("  Consider regenerating with a passphrase for stronger protection.\n\n")
	b.WriteString(helpStyle.Render("  enter:I understand, continue  esc:back to home"))
	return b.String()
}

func (g *GenKey) viewDone() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Key generated"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s       (0600)\n", sshDisplayPath(g.homeDir, g.sshDir, g.name)))
	b.WriteString(fmt.Sprintf("  %s   (0644)\n", sshDisplayPath(g.homeDir, g.sshDir, g.name)+".pub"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  enter/esc:back to home"))
	return b.String()
}

func (g *GenKey) viewError() string {
	var b strings.Builder
	b.WriteString(warningStyle.Render("  Generation failed"))
	b.WriteString("\n\n")
	if g.err != nil {
		b.WriteString(fmt.Sprintf("  %s\n\n", g.err.Error()))
	}
	b.WriteString(helpStyle.Render("  enter/esc:back to home"))
	return b.String()
}

func keyTypeLabel(kt keys.KeyType) string {
	switch kt {
	case keys.KeyTypeRSA4096:
		return "rsa4096"
	default:
		return "ed25519"
	}
}

func sshDisplayPath(homeDir, sshDir, name string) string {
	if homeDir != "" {
		expected := filepath.Join(homeDir, ".ssh")
		if sshDir == expected {
			if name == "" {
				return "~/.ssh/"
			}
			return "~/.ssh/" + name
		}
	}
	if name == "" {
		return sshDir + string(os.PathSeparator)
	}
	return filepath.Join(sshDir, name)
}
