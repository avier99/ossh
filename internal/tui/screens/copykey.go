package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinburke/ssh_config"
	"github.com/synoxisllc/ossh/internal/config"
	"github.com/synoxisllc/ossh/internal/keys"
	"github.com/synoxisllc/ossh/internal/tui"
)

type copyKeyStep int

const (
	stepCKKeySelect copyKeyStep = iota
	stepCKConfirm
	stepCKDone
	stepCKError
)

// CopyKey is the copy-key-to-server wizard screen.
type CopyKey struct {
	step       copyKeyStep
	sshDir     string
	cfg        *ssh_config.Config
	findings   []config.Finding
	host       hostItem
	keyEntries []keys.KeyEntry
	keyChoice  int
	err        error
	fallback   string
}

// NewCopyKey creates a new copy-key wizard for the given host.
func NewCopyKey(sshDir string, cfg *ssh_config.Config, findings []config.Finding, host hostItem) *CopyKey {
	keyEntries, _ := keys.ListPrivateKeys(sshDir)
	return &CopyKey{
		step:       stepCKKeySelect,
		sshDir:     sshDir,
		cfg:        cfg,
		findings:   findings,
		host:       host,
		keyEntries: keyEntries,
		keyChoice:  0,
	}
}

// Init implements tui.Screen.
func (g *CopyKey) Init() tea.Cmd {
	return nil
}

// Update implements tui.Screen.
func (g *CopyKey) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case copyKeyFinishedMsg:
		return g.handleFinished(msg)
	case tea.KeyMsg:
		return g.handleKey(msg)
	}
	return g, nil
}

type copyKeyFinishedMsg struct {
	err error
}

func (g *CopyKey) handleFinished(msg copyKeyFinishedMsg) (tui.Screen, tea.Cmd) {
	if msg.err != nil {
		g.step = stepCKError
		g.err = msg.err
		return g, nil
	}
	g.step = stepCKDone
	return g, nil
}

func (g *CopyKey) handleKey(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch g.step {
	case stepCKKeySelect:
		return g.handleCKKeySelect(msg)
	case stepCKConfirm:
		return g.handleCKConfirm(msg)
	case stepCKDone:
		return g.handleCKDone(msg)
	case stepCKError:
		return g.handleCKError(msg)
	}
	return g, nil
}

func (g *CopyKey) homeScreen() *Home {
	return NewHome(g.cfg, g.findings, g.sshDir)
}

func (g *CopyKey) copyOpts() keys.CopyKeyOpts {
	return keys.CopyKeyOpts{
		SSHDir:    g.sshDir,
		Key:       g.keyEntries[g.keyChoice],
		HostAlias: g.host.Alias,
		User:      g.host.User,
	}
}

func (g *CopyKey) handleCKKeySelect(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	if len(g.keyEntries) == 0 {
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

	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		return g, func() tea.Msg {
			return tui.SwitchScreenMsg{Next: g.homeScreen()}
		}
	case "up", "k":
		if g.keyChoice > 0 {
			g.keyChoice--
		}
		return g, nil
	case "down", "j":
		if g.keyChoice < len(g.keyEntries)-1 {
			g.keyChoice++
		}
		return g, nil
	case "enter":
		opts := g.copyOpts()
		g.fallback = keys.ManualFallback(opts)
		g.step = stepCKConfirm
		return g, nil
	}
	return g, nil
}

func (g *CopyKey) handleCKConfirm(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepCKKeySelect
		return g, nil
	case "enter":
		opts := g.copyOpts()
		cmd, err := keys.CopyCommand(opts)
		if err != nil {
			g.err = err
			g.step = stepCKError
			return g, nil
		}
		return g, tea.ExecProcess(cmd, func(execErr error) tea.Msg {
			return copyKeyFinishedMsg{err: execErr}
		})
	}
	return g, nil
}

func (g *CopyKey) handleCKDone(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
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

func (g *CopyKey) handleCKError(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
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
func (g *CopyKey) View() string {
	switch g.step {
	case stepCKKeySelect:
		return g.viewCKKeySelect()
	case stepCKConfirm:
		return g.viewCKConfirm()
	case stepCKDone:
		return g.viewCKDone()
	case stepCKError:
		return g.viewCKError()
	}
	return ""
}

func (g *CopyKey) viewCKKeySelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("  Copy key to %s", g.host.Alias)))
	b.WriteString("\n\n")

	if len(g.keyEntries) == 0 {
		b.WriteString("  No keys found in ~/.ssh/.\n")
		b.WriteString("  Generate one with g on the home screen.\n\n")
		b.WriteString(helpStyle.Render("  enter/esc:back"))
		return b.String()
	}

	b.WriteString("  Select key:\n\n")
	for i, k := range g.keyEntries {
		prefix := "    "
		if i == g.keyChoice {
			prefix = "  ▶ "
		}
		line := prefix + k.DisplayPath
		if i == g.keyChoice {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter:select  ↑↓:navigate  esc:cancel"))
	return b.String()
}

func (g *CopyKey) viewCKConfirm() string {
	opts := g.copyOpts()
	cmdLine := fmt.Sprintf("ssh-copy-id -i %s %s", opts.Key.DisplayPath+".pub", keys.Target(opts))

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("  Copy key to %s", g.host.Alias)))
	b.WriteString("\n\n")
	b.WriteString("  Will run:\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", cmdLine))
	b.WriteString("  You will be prompted for the remote password.\n\n")
	b.WriteString(helpStyle.Render("  enter:run  esc:back"))
	return b.String()
}

func (g *CopyKey) viewCKDone() string {
	opts := g.copyOpts()
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Key copied"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s.pub → %s\n", opts.Key.DisplayPath, keys.Target(opts)))
	b.WriteString("  Public key is now in the server's authorized_keys.\n\n")
	b.WriteString(helpStyle.Render("  enter/esc:back to home"))
	return b.String()
}

func (g *CopyKey) viewCKError() string {
	var b strings.Builder
	b.WriteString(warningStyle.Render("  [!] Copy failed"))
	b.WriteString("\n\n")
	if g.err != nil {
		b.WriteString(fmt.Sprintf("  %s\n\n", g.err.Error()))
	}
	b.WriteString("  Manual fallback:\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", g.fallback))
	b.WriteString(helpStyle.Render("  enter/esc:back to home"))
	return b.String()
}
