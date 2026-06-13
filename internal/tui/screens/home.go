package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevinburke/ssh_config"
	"github.com/sahilm/fuzzy"
	"github.com/synoxisllc/ossh/internal/config"
	"github.com/synoxisllc/ossh/internal/tui"
)

// hostItem holds the data for a single SSH host entry
type hostItem struct {
	Alias    string
	Hostname string
	User     string
}

// Home is the main screen showing fuzzy-searchable host list
type Home struct {
	hosts        []hostItem
	query        string
	matches      []fuzzy.Match
	cursor       int
	findingCount int
	sshDir       string
	cfg          *ssh_config.Config
	findings     []config.Finding
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	queryStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	emptyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
)

// NewHome creates a new home screen from the SSH config and audit findings
func NewHome(cfg *ssh_config.Config, findings []config.Finding, sshDir string) *Home {
	hosts := extractHosts(cfg)

	return &Home{
		hosts:        hosts,
		query:        "",
		matches:      nil, // nil means show all
		cursor:       0,
		findingCount: len(findings),
		sshDir:       sshDir,
		cfg:          cfg,
		findings:     findings,
	}
}

// extractHosts pulls all Host entries from the SSH config
func extractHosts(cfg *ssh_config.Config) []hostItem {
	if cfg == nil {
		return nil
	}

	var hosts []hostItem
	for _, host := range cfg.Hosts {
		// Skip wildcard patterns and special entries
		for _, pattern := range host.Patterns {
			if pattern.String() == "*" || strings.Contains(pattern.String(), "*") {
				continue
			}

			alias := pattern.String()
			hostname, _ := cfg.Get(alias, "HostName")
			user, _ := cfg.Get(alias, "User")

			// If HostName is not set, it defaults to the alias
			if hostname == "" {
				hostname = alias
			}

			hosts = append(hosts, hostItem{
				Alias:    alias,
				Hostname: hostname,
				User:     user,
			})
			break // Only take first non-wildcard pattern per host
		}
	}

	return hosts
}

// Init implements tui.Screen
func (h *Home) Init() tea.Cmd {
	return nil
}

// Update implements tui.Screen
func (h *Home) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return h.handleKey(msg)
	}
	return h, nil
}

func (h *Home) handleKey(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return h, tea.Quit

	case "esc":
		if h.query != "" {
			// Clear query
			h.query = ""
			h.matches = nil
			h.cursor = 0
			return h, nil
		}
		// Empty query — quit
		return h, tea.Quit

	case "enter":
		// Connect to selected host
		selected := h.getSelectedHost()
		if selected != nil {
			return h, func() tea.Msg {
				return tui.ConnectMsg{HostAlias: selected.Alias}
			}
		}
		return h, nil

	case "up", "k":
		if h.cursor > 0 {
			h.cursor--
		}
		return h, nil

	case "down", "j":
		maxIdx := h.getResultCount() - 1
		if h.cursor < maxIdx {
			h.cursor++
		}
		return h, nil

	case "g":
		if h.query == "" {
			return h, func() tea.Msg {
				return tui.SwitchScreenMsg{Next: NewGenKey(h.sshDir, h.cfg, h.findings)}
			}
		}
		h.query += "g"
		h.runFuzzy()
		return h, nil

	case "backspace":
		if len(h.query) > 0 {
			h.query = h.query[:len(h.query)-1]
			h.runFuzzy()
			return h, nil
		}
		return h, nil

	default:
		// Append printable characters to query
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			if r >= 32 && r < 127 { // Printable ASCII
				h.query += string(r)
				h.runFuzzy()
				return h, nil
			}
		}
		return h, nil
	}
}

func (h *Home) runFuzzy() {
	if h.query == "" {
		h.matches = nil
		h.cursor = 0
		return
	}

	h.matches = fuzzy.FindFrom(h.query, h)
	h.cursor = 0
}

func (h *Home) getResultCount() int {
	if h.matches == nil {
		return len(h.hosts)
	}
	return len(h.matches)
}

func (h *Home) getSelectedHost() *hostItem {
	if h.matches == nil {
		// No filter — return from hosts directly
		if h.cursor >= 0 && h.cursor < len(h.hosts) {
			return &h.hosts[h.cursor]
		}
	} else {
		// Filtered results
		if h.cursor >= 0 && h.cursor < len(h.matches) {
			idx := h.matches[h.cursor].Index
			if idx >= 0 && idx < len(h.hosts) {
				return &h.hosts[idx]
			}
		}
	}
	return nil
}

// View implements tui.Screen
func (h *Home) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("  ossh"))
	b.WriteString("\n\n")

	// Empty state
	if len(h.hosts) == 0 {
		b.WriteString(emptyStyle.Render("  No hosts configured yet. Use Add Host to get started."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  ctrl+c:quit"))
		return b.String()
	}

	// Query
	b.WriteString("  > ")
	b.WriteString(queryStyle.Render(h.query))
	b.WriteString("\n\n")

	// Results
	if h.matches != nil && len(h.matches) == 0 {
		// No matches for query
		b.WriteString(emptyStyle.Render(fmt.Sprintf("  No matches for %q", h.query)))
	} else {
		h.renderResults(&b)
	}

	b.WriteString("\n\n")

	// Warning if findings exist
	if h.findingCount > 0 {
		warning := fmt.Sprintf("[!] %d permission issue", h.findingCount)
		if h.findingCount > 1 {
			warning += "s"
		}
		warning += " found"
		b.WriteString(warningStyle.Render("  " + warning))
		b.WriteString("\n\n")
	}

	// Help
	b.WriteString(helpStyle.Render("  enter:connect  ↑↓:navigate  esc:clear/quit"))

	return b.String()
}

func (h *Home) renderResults(b *strings.Builder) {
	results := h.getResults()
	for i, item := range results {
		prefix := "  "
		if i == h.cursor {
			prefix = "▶ "
		}

		line := fmt.Sprintf("%s%-15s  %-20s  %s",
			prefix,
			item.Alias,
			item.Hostname,
			item.User,
		)

		if i == h.cursor {
			line = selectedStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

func (h *Home) getResults() []hostItem {
	if h.matches == nil {
		return h.hosts
	}

	results := make([]hostItem, len(h.matches))
	for i, match := range h.matches {
		results[i] = h.hosts[match.Index]
	}
	return results
}

// Implement fuzzy.Source interface so fuzzy matching works across all fields

func (h *Home) String(i int) string {
	if i < 0 || i >= len(h.hosts) {
		return ""
	}
	item := h.hosts[i]
	// Concatenate all searchable fields
	return item.Alias + " " + item.Hostname + " " + item.User
}

func (h *Home) Len() int {
	return len(h.hosts)
}
