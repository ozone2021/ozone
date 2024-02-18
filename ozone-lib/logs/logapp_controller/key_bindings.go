package logapp_controller

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Follow       key.Binding
	NextLog      key.Binding
	PreviousLog  key.Binding
	PageDown     key.Binding
	PageUp       key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	Down         key.Binding
	Up           key.Binding
}

func LogKeyMap() KeyMap {
	return KeyMap{
		Follow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow logs"),
		),
		NextLog: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next log"),
		),
		PreviousLog: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "previous log"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("b/pgup", "page up"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
	}
}
