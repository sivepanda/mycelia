package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for TUI navigation.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Select key.Binding
	Back   key.Binding
	Manual key.Binding
	Quit   key.Binding
	Help   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Select, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Left, k.Right},
		{k.Select, k.Enter, k.Back, k.Manual, k.Help, k.Quit},
	}
}

// DefaultKeys returns the default key bindings.
func DefaultKeys() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "continue"),
		),
		Select: key.NewBinding(
			key.WithKeys("x", " "),
			key.WithHelp("x/space", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Manual: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "add manual"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}
