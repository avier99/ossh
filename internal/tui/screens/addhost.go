package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinburke/ssh_config"
	"github.com/synoxisllc/ossh/internal/config"
	"github.com/synoxisllc/ossh/internal/keys"
	"github.com/synoxisllc/ossh/internal/tui"
)

type addHostStep int

const (
	stepAHAlias addHostStep = iota
	stepAHHostname
	stepAHUser
	stepAHPort
	stepAHAuthMethod
	stepAHKeyFile
	stepAHPasswordWarn
	stepAHConfirm
	stepAHDone
	stepAHError
)

// AddHost is the add-host wizard screen.
type AddHost struct {
	step          addHostStep
	sshDir        string
	homeDir       string
	cfg           *ssh_config.Config
	findings      []config.Finding
	aliasInput    textinput.Model
	hostnameInput textinput.Model
	userInput     textinput.Model
	portInput     textinput.Model
	authChoice    int
	keyEntries    []keys.KeyEntry
	keyChoice     int
	alias         string
	hostname      string
	user          string
	port          int
	authMethod    string
	identityFile  string
	aliasErr      error
	hostnameErr   error
	portErr       error
	err           error
	newCfg        *ssh_config.Config
	newFindings   []config.Finding
}

// NewAddHost creates a new add-host wizard.
func NewAddHost(sshDir string, cfg *ssh_config.Config, findings []config.Finding) *AddHost {
	aliasInput := textinput.New()
	aliasInput.Placeholder = "homelab"
	aliasInput.CharLimit = 64
	aliasInput.Width = 30
	aliasInput.Focus()

	hostnameInput := textinput.New()
	hostnameInput.Placeholder = "192.168.1.100"
	hostnameInput.CharLimit = 255
	hostnameInput.Width = 30

	userInput := textinput.New()
	userInput.Placeholder = "gana"
	userInput.CharLimit = 64
	userInput.Width = 30

	portInput := textinput.New()
	portInput.Placeholder = "22"
	portInput.CharLimit = 5
	portInput.Width = 10

	homeDir, _ := os.UserHomeDir()
	keyEntries, _ := keys.ListPrivateKeys(sshDir)

	return &AddHost{
		step:          stepAHAlias,
		sshDir:        sshDir,
		homeDir:       homeDir,
		cfg:           cfg,
		findings:      findings,
		aliasInput:    aliasInput,
		hostnameInput: hostnameInput,
		userInput:     userInput,
		portInput:     portInput,
		authChoice:    0,
		keyEntries:    keyEntries,
		keyChoice:     0,
	}
}

// Init implements tui.Screen.
func (g *AddHost) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tui.Screen.
func (g *AddHost) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return g.handleKey(msg)
	}

	var cmd tea.Cmd
	switch g.step {
	case stepAHAlias:
		g.aliasInput, cmd = g.aliasInput.Update(msg)
	case stepAHHostname:
		g.hostnameInput, cmd = g.hostnameInput.Update(msg)
	case stepAHUser:
		g.userInput, cmd = g.userInput.Update(msg)
	case stepAHPort:
		g.portInput, cmd = g.portInput.Update(msg)
	}
	return g, cmd
}

func (g *AddHost) handleKey(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch g.step {
	case stepAHAlias:
		return g.handleAHAlias(msg)
	case stepAHHostname:
		return g.handleAHHostname(msg)
	case stepAHUser:
		return g.handleAHUser(msg)
	case stepAHPort:
		return g.handleAHPort(msg)
	case stepAHAuthMethod:
		return g.handleAHAuthMethod(msg)
	case stepAHKeyFile:
		return g.handleAHKeyFile(msg)
	case stepAHPasswordWarn:
		return g.handleAHPasswordWarn(msg)
	case stepAHConfirm:
		return g.handleAHConfirm(msg)
	case stepAHDone:
		return g.handleAHDone(msg)
	case stepAHError:
		return g.handleAHError(msg)
	}
	return g, nil
}

func (g *AddHost) homeScreen() *Home {
	if g.newCfg != nil {
		return NewHome(g.newCfg, g.newFindings, g.sshDir)
	}
	return NewHome(g.cfg, g.findings, g.sshDir)
}

func (g *AddHost) handleAHAlias(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		return g, func() tea.Msg {
			return tui.SwitchScreenMsg{Next: g.homeScreen()}
		}
	case "enter":
		alias := strings.TrimSpace(g.aliasInput.Value())
		if err := config.ValidateAlias(alias, g.cfg); err != nil {
			g.aliasErr = err
			return g, nil
		}
		g.alias = alias
		g.aliasErr = nil
		g.step = stepAHHostname
		g.aliasInput.Blur()
		g.hostnameInput.Focus()
		return g, nil
	}

	var cmd tea.Cmd
	g.aliasInput, cmd = g.aliasInput.Update(msg)
	return g, cmd
}

func (g *AddHost) handleAHHostname(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepAHAlias
		g.hostnameErr = nil
		g.hostnameInput.Blur()
		g.aliasInput.Focus()
		return g, nil
	case "enter":
		hostname := strings.TrimSpace(g.hostnameInput.Value())
		if hostname == "" {
			g.hostnameErr = fmt.Errorf("hostname cannot be empty")
			return g, nil
		}
		g.hostname = hostname
		g.hostnameErr = nil
		g.step = stepAHUser
		g.hostnameInput.Blur()
		g.userInput.Focus()
		return g, nil
	}

	var cmd tea.Cmd
	g.hostnameInput, cmd = g.hostnameInput.Update(msg)
	return g, cmd
}

func (g *AddHost) handleAHUser(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepAHHostname
		g.userInput.Blur()
		g.hostnameInput.Focus()
		return g, nil
	case "enter":
		g.user = strings.TrimSpace(g.userInput.Value())
		g.step = stepAHPort
		g.userInput.Blur()
		g.portInput.Focus()
		return g, nil
	}

	var cmd tea.Cmd
	g.userInput, cmd = g.userInput.Update(msg)
	return g, cmd
}

func (g *AddHost) handleAHPort(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepAHUser
		g.portErr = nil
		g.portInput.Blur()
		g.userInput.Focus()
		return g, nil
	case "enter":
		portStr := strings.TrimSpace(g.portInput.Value())
		if portStr == "" {
			g.port = 0
			g.portErr = nil
			g.step = stepAHAuthMethod
			g.portInput.Blur()
			return g, nil
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			g.portErr = fmt.Errorf("port must be between 1 and 65535")
			return g, nil
		}
		g.port = port
		g.portErr = nil
		g.step = stepAHAuthMethod
		g.portInput.Blur()
		return g, nil
	}

	var cmd tea.Cmd
	g.portInput, cmd = g.portInput.Update(msg)
	return g, cmd
}

func (g *AddHost) handleAHAuthMethod(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepAHPort
		g.portInput.Focus()
		return g, nil
	case "up", "k":
		if g.authChoice > 0 {
			g.authChoice--
		}
		return g, nil
	case "down", "j":
		if g.authChoice < 1 {
			g.authChoice++
		}
		return g, nil
	case "enter":
		if g.authChoice == 0 {
			g.authMethod = "key"
			g.step = stepAHKeyFile
		} else {
			g.authMethod = "password"
			g.step = stepAHPasswordWarn
		}
		return g, nil
	}
	return g, nil
}

func (g *AddHost) handleAHKeyFile(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	maxChoice := len(g.keyEntries) // index 0 = skip, 1..n = keys

	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepAHAuthMethod
		return g, nil
	case "up", "k":
		if g.keyChoice > 0 {
			g.keyChoice--
		}
		return g, nil
	case "down", "j":
		if g.keyChoice < maxChoice {
			g.keyChoice++
		}
		return g, nil
	case "enter":
		if g.keyChoice == 0 {
			g.identityFile = ""
		} else {
			g.identityFile = g.keyEntries[g.keyChoice-1].DisplayPath
		}
		g.step = stepAHConfirm
		return g, nil
	}
	return g, nil
}

func (g *AddHost) handleAHPasswordWarn(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		g.step = stepAHAuthMethod
		return g, nil
	case "enter":
		g.step = stepAHConfirm
		return g, nil
	}
	return g, nil
}

func (g *AddHost) handleAHConfirm(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return g, tea.Quit
	case "esc":
		if g.authMethod == "key" {
			g.step = stepAHKeyFile
		} else {
			g.step = stepAHPasswordWarn
		}
		return g, nil
	case "enter":
		opts := config.AddHostOpts{
			Alias:        g.alias,
			Hostname:     g.hostname,
			User:         g.user,
			Port:         g.port,
			AuthMethod:   g.authMethod,
			IdentityFile: g.identityFile,
		}
		if err := config.AddHost(g.sshDir, opts); err != nil {
			g.step = stepAHError
			g.err = err
			return g, nil
		}
		newCfg, _ := config.Load(filepath.Join(g.sshDir, "config"))
		newFindings, _ := config.Audit(g.sshDir)
		g.newCfg = newCfg
		g.newFindings = newFindings
		g.step = stepAHDone
		return g, nil
	}
	return g, nil
}

func (g *AddHost) handleAHDone(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
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

func (g *AddHost) handleAHError(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
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
func (g *AddHost) View() string {
	switch g.step {
	case stepAHAlias:
		return g.viewAHAlias()
	case stepAHHostname:
		return g.viewAHHostname()
	case stepAHUser:
		return g.viewAHUser()
	case stepAHPort:
		return g.viewAHPort()
	case stepAHAuthMethod:
		return g.viewAHAuthMethod()
	case stepAHKeyFile:
		return g.viewAHKeyFile()
	case stepAHPasswordWarn:
		return g.viewAHPasswordWarn()
	case stepAHConfirm:
		return g.viewAHConfirm()
	case stepAHDone:
		return g.viewAHDone()
	case stepAHError:
		return g.viewAHError()
	}
	return ""
}

func (g *AddHost) viewAHAlias() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString("  Alias: ")
	b.WriteString(g.aliasInput.View())
	b.WriteString("\n\n")
	if g.aliasErr != nil {
		b.WriteString(warningStyle.Render("  " + g.aliasErr.Error()))
		b.WriteString("\n\n")
	}
	b.WriteString(helpStyle.Render("  enter:next  esc:cancel"))
	return b.String()
}

func (g *AddHost) viewAHHostname() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString("  Hostname / IP: ")
	b.WriteString(g.hostnameInput.View())
	b.WriteString("\n\n")
	if g.hostnameErr != nil {
		b.WriteString(warningStyle.Render("  " + g.hostnameErr.Error()))
		b.WriteString("\n\n")
	}
	b.WriteString(helpStyle.Render("  enter:next  esc:back"))
	return b.String()
}

func (g *AddHost) viewAHUser() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString("  User (leave blank for system default): ")
	b.WriteString(g.userInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  enter:next  esc:back"))
	return b.String()
}

func (g *AddHost) viewAHPort() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString("  Port (leave blank for 22): ")
	b.WriteString(g.portInput.View())
	b.WriteString("\n\n")
	if g.portErr != nil {
		b.WriteString(warningStyle.Render("  " + g.portErr.Error()))
		b.WriteString("\n\n")
	}
	b.WriteString(helpStyle.Render("  enter:next  esc:back"))
	return b.String()
}

func (g *AddHost) viewAHAuthMethod() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString("  Authentication method:\n\n")

	options := []struct {
		label string
		note  string
	}{
		{"Key", "(recommended)"},
		{"Password", ""},
	}
	for i, opt := range options {
		prefix := "    "
		if i == g.authChoice {
			prefix = "  ▶ "
		}
		line := opt.label
		if opt.note != "" {
			line = fmt.Sprintf("%-12s  %s", opt.label, opt.note)
		}
		line = prefix + line
		if i == g.authChoice {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter:select  ↑↓:choose  esc:back"))
	return b.String()
}

func (g *AddHost) viewAHKeyFile() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString("  Identity file:\n\n")

	items := make([]string, 0, 1+len(g.keyEntries))
	items = append(items, "(none — skip)")
	for _, k := range g.keyEntries {
		items = append(items, k.DisplayPath)
	}

	for i, item := range items {
		prefix := "    "
		if i == g.keyChoice {
			prefix = "  ▶ "
		}
		line := prefix + item
		if i == g.keyChoice {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter:select  ↑↓:navigate  esc:back"))
	return b.String()
}

func (g *AddHost) viewAHPasswordWarn() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString(warningStyle.Render("  [!] Password authentication"))
	b.WriteString("\n\n")
	b.WriteString("  Password auth sends your credentials over the network on\n")
	b.WriteString("  every login. They can be intercepted, logged, or brute-forced.\n\n")
	b.WriteString("  Key-based auth is strongly recommended. Only continue if you\n")
	b.WriteString("  understand the risk and have no alternative.\n\n")
	b.WriteString(helpStyle.Render("  enter:I understand, continue  esc:back"))
	return b.String()
}

func (g *AddHost) viewAHConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add host"))
	b.WriteString("\n\n")
	b.WriteString("  Confirm:\n\n")
	b.WriteString(fmt.Sprintf("  Alias:    %s\n", g.alias))
	b.WriteString(fmt.Sprintf("  Hostname: %s\n", g.hostname))
	if g.user != "" {
		b.WriteString(fmt.Sprintf("  User:     %s\n", g.user))
	} else {
		b.WriteString("  User:     (system default)\n")
	}
	portLabel := "22 (default)"
	if g.port != 0 {
		portLabel = strconv.Itoa(g.port)
	}
	b.WriteString(fmt.Sprintf("  Port:     %s\n", portLabel))
	if g.authMethod == "key" {
		b.WriteString("  Auth:     key\n")
		if g.identityFile != "" {
			b.WriteString(fmt.Sprintf("  Key:      %s\n", g.identityFile))
		} else {
			b.WriteString("  Key:      (none)\n")
		}
	} else {
		b.WriteString("  Auth:     password\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter:add host  esc:back"))
	return b.String()
}

func (g *AddHost) viewAHDone() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Host added"))
	b.WriteString("\n\n")
	configDisplay := filepath.Join(g.sshDir, "config")
	if g.homeDir != "" && g.sshDir == filepath.Join(g.homeDir, ".ssh") {
		configDisplay = "~/.ssh/config"
	}
	b.WriteString(fmt.Sprintf("  %s added to %s\n", g.alias, configDisplay))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter/esc:back to home"))
	return b.String()
}

func (g *AddHost) viewAHError() string {
	var b strings.Builder
	b.WriteString(warningStyle.Render("  Add host failed"))
	b.WriteString("\n\n")
	if g.err != nil {
		b.WriteString(fmt.Sprintf("  %s\n\n", g.err.Error()))
	}
	b.WriteString(helpStyle.Render("  enter/esc:back to home"))
	return b.String()
}
