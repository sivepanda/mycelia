package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormStyles configures the visual appearance of a Form.
type FormStyles struct {
	LabelStyle       lipgloss.Style
	ActiveLabelStyle lipgloss.Style
	InputWidth       int
}

// DefaultFormStyles returns sensible defaults.
func DefaultFormStyles() FormStyles {
	return FormStyles{
		LabelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C0CAF5")),
		ActiveLabelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9E64")).
			Bold(true),
		InputWidth: 50,
	}
}

// FormField describes a single field in the form.
type FormField struct {
	Key         string // Identifier used in Values()
	Label       string // Display label
	Placeholder string
	CharLimit   int
}

// Form is a headless multi-field text input primitive.
// It never handles tea.KeyMsg — the parent drives it via imperative methods.
type Form struct {
	fields []FormField
	inputs []textinput.Model
	focus  int
	styles FormStyles
}

// NewForm creates a Form from field definitions.
func NewForm(fields []FormField) Form {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.CharLimit = f.CharLimit
		if ti.CharLimit == 0 {
			ti.CharLimit = 256
		}
		ti.Width = 50
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return Form{
		fields: fields,
		inputs: inputs,
		focus:  0,
		styles: DefaultFormStyles(),
	}
}

// WithStyles replaces the style config.
func (f Form) WithStyles(styles FormStyles) Form {
	f.styles = styles
	if styles.InputWidth > 0 {
		for i := range f.inputs {
			f.inputs[i].Width = styles.InputWidth
		}
	}
	return f
}

// --- Imperative control ---

// NextField moves focus to the next field (wraps around).
func (f Form) NextField() Form {
	f.inputs[f.focus].Blur()
	f.focus = (f.focus + 1) % len(f.inputs)
	f.inputs[f.focus].Focus()
	return f
}

// PrevField moves focus to the previous field (wraps around).
func (f Form) PrevField() Form {
	f.inputs[f.focus].Blur()
	f.focus = (f.focus - 1 + len(f.inputs)) % len(f.inputs)
	f.inputs[f.focus].Focus()
	return f
}

// UpdateInput forwards a tea.Msg to the currently focused input.
// This is how the parent passes key events to the text input.
func (f Form) UpdateInput(msg tea.Msg) (Form, tea.Cmd) {
	var cmd tea.Cmd
	f.inputs[f.focus], cmd = f.inputs[f.focus].Update(msg)
	return f, cmd
}

// Reset clears all field values and returns focus to the first field.
func (f Form) Reset() Form {
	for i := range f.inputs {
		f.inputs[i].SetValue("")
		if i == 0 {
			f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
	}
	f.focus = 0
	return f
}

// --- State inspection ---

// Values returns a map of field keys to their current values.
func (f Form) Values() map[string]string {
	vals := make(map[string]string, len(f.fields))
	for i, field := range f.fields {
		vals[field.Key] = f.inputs[i].Value()
	}
	return vals
}

// Value returns the current value for a specific field key.
func (f Form) Value(key string) string {
	for i, field := range f.fields {
		if field.Key == key {
			return f.inputs[i].Value()
		}
	}
	return ""
}

// FocusedField returns the index of the currently focused field.
func (f Form) FocusedField() int {
	return f.focus
}

// --- Rendering ---

// View renders the form fields.
func (f Form) View() string {
	var lines []string
	for i, field := range f.fields {
		label := field.Label
		if i == f.focus {
			label = f.styles.ActiveLabelStyle.Render(label)
		} else {
			label = f.styles.LabelStyle.Render(label)
		}
		lines = append(lines, label)
		lines = append(lines, "  "+f.inputs[i].View())
		if i < len(f.fields)-1 {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}
