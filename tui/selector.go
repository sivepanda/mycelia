package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SelectorStyles configures the visual appearance of a Selector.
type SelectorStyles struct {
	Cursor       string
	CursorBlank  string
	Checked      string
	Unchecked    string
	ItemStyle    lipgloss.Style
	ActiveStyle  lipgloss.Style
	DetailStyle  lipgloss.Style
	ActiveDetail lipgloss.Style
}

// DefaultSelectorStyles returns sensible defaults using Tokyo Night colors.
func DefaultSelectorStyles() SelectorStyles {
	return SelectorStyles{
		Cursor:      "❯ ",
		CursorBlank: "  ",
		Checked:     "☑",
		Unchecked:   "☐",
		ItemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C0CAF5")),
		ActiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9E64")).
			Bold(true),
		DetailStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#414868")).
			PaddingLeft(4),
		ActiveDetail: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89")).
			PaddingLeft(4),
	}
}

// SelectorItem wraps an item with its selection state.
type SelectorItem[T any] struct {
	Value    T
	Selected bool
}

// Selector is a headless multi-select list primitive.
// It never handles tea.KeyMsg — the parent drives it via imperative methods.
type Selector[T any] struct {
	items      []SelectorItem[T]
	cursor     int
	labelFunc  func(T) string
	detailFunc func(T) string
	styles     SelectorStyles
	width      int
	align      lipgloss.Position
}

// NewSelector creates a Selector. All items start unselected.
func NewSelector[T any](items []T, labelFunc func(T) string) Selector[T] {
	wrapped := make([]SelectorItem[T], len(items))
	for i, item := range items {
		wrapped[i] = SelectorItem[T]{Value: item}
	}
	return Selector[T]{
		items:     wrapped,
		labelFunc: labelFunc,
		styles:    DefaultSelectorStyles(),
		width:     60,
	}
}

// WithStyles replaces the style config.
func (s Selector[T]) WithStyles(styles SelectorStyles) Selector[T] {
	s.styles = styles
	return s
}

// WithDetail adds an optional detail line rendered below each item.
func (s Selector[T]) WithDetail(fn func(T) string) Selector[T] {
	s.detailFunc = fn
	return s
}

// WithSelected sets initial selection state for all items.
func (s Selector[T]) WithSelected(selected bool) Selector[T] {
	for i := range s.items {
		s.items[i].Selected = selected
	}
	return s
}

// WithWidth sets the render width.
func (s Selector[T]) WithWidth(w int) Selector[T] {
	s.width = w
	return s
}

// WithAlign sets the horizontal alignment for rendered content.
func (s Selector[T]) WithAlign(align lipgloss.Position) Selector[T] {
	s.align = align
	return s
}

// --- Imperative control ---

// CursorUp moves the cursor up one position.
func (s Selector[T]) CursorUp() Selector[T] {
	if s.cursor > 0 {
		s.cursor--
	}
	return s
}

// CursorDown moves the cursor down one position.
func (s Selector[T]) CursorDown() Selector[T] {
	if s.cursor < len(s.items)-1 {
		s.cursor++
	}
	return s
}

// Toggle flips the selection state of the item under the cursor.
func (s Selector[T]) Toggle() Selector[T] {
	if len(s.items) > 0 {
		s.items[s.cursor].Selected = !s.items[s.cursor].Selected
	}
	return s
}

// --- State inspection ---

// Cursor returns the current cursor position.
func (s Selector[T]) Cursor() int {
	return s.cursor
}

// Selected returns all selected item values.
func (s Selector[T]) Selected() []T {
	var out []T
	for _, item := range s.items {
		if item.Selected {
			out = append(out, item.Value)
		}
	}
	return out
}

// Items returns all items with their selection state.
func (s Selector[T]) Items() []SelectorItem[T] {
	return s.items
}

// HasSelected returns true if at least one item is selected.
func (s Selector[T]) HasSelected() bool {
	for _, item := range s.items {
		if item.Selected {
			return true
		}
	}
	return false
}

// Len returns the number of items.
func (s Selector[T]) Len() int {
	return len(s.items)
}

// SetItems replaces the item list and resets the cursor.
func (s Selector[T]) SetItems(items []T) Selector[T] {
	wrapped := make([]SelectorItem[T], len(items))
	for i, item := range items {
		wrapped[i] = SelectorItem[T]{Value: item}
	}
	s.items = wrapped
	s.cursor = 0
	return s
}

// --- Rendering ---

// View renders the selector list.
func (s Selector[T]) View() string {
	if len(s.items) == 0 {
		return s.styles.ItemStyle.Render("  (no items)")
	}

	var lines []string
	for i, item := range s.items {
		cursor := s.styles.CursorBlank
		if i == s.cursor {
			cursor = s.styles.Cursor
		}

		checkbox := s.styles.Unchecked
		if item.Selected {
			checkbox = s.styles.Checked
		}

		label := s.labelFunc(item.Value)
		line := fmt.Sprintf("%s%s %s", cursor, checkbox, label)

		if i == s.cursor {
			line = s.styles.ActiveStyle.Render(line)
		} else {
			line = s.styles.ItemStyle.Render(line)
		}
		lines = append(lines, line)

		// Optional detail line
		if s.detailFunc != nil {
			detail := s.detailFunc(item.Value)
			if detail != "" {
				if i == s.cursor {
					lines = append(lines, s.styles.ActiveDetail.Render(detail))
				} else {
					lines = append(lines, s.styles.DetailStyle.Render(detail))
				}
			}
		}
	}

	content := strings.Join(lines, "\n")
	if s.width > 0 {
		return lipgloss.PlaceHorizontal(s.width, s.align, content)
	}
	return content
}
