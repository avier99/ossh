package tui

// SwitchScreenMsg signals the root model to switch to a new screen.
type SwitchScreenMsg struct {
	Next Screen
}

// ConnectMsg signals a request to connect to a host via SSH.
// The root model handles this by calling ssh.Connect and quitting the TUI.
type ConnectMsg struct {
	HostAlias string
}
