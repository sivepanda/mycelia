package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sivepanda/mycelia"
	primitives "github.com/sivepanda/mycelia/tui"
)

// MenuAction identifies a top-level menu option.
type MenuAction int

const (
	ActionSetup MenuAction = iota
	ActionDetect
	ActionList
	ActionSuggest
	ActionAuto
	ActionEdit
)

type menuItem struct {
	action MenuAction
	label  string
	desc   string
	icon   string
}

var menuItems = []menuItem{
	{ActionSetup, "Setup", "Interactive housekeeping wizard", IconSettings},
	{ActionDetect, "Detect", "Detect package managers", IconInfo},
	{ActionList, "List", "List configured commands", IconBullet},
	{ActionSuggest, "Suggest", "Show suggested commands", IconArrow},
	{ActionAuto, "Auto", "Auto-detect and add commands", IconCheck},
	{ActionEdit, "Edit", "Open config in $EDITOR", IconSettings},
}

// MenuView represents whether we're on the main menu or inside a sub-view.
type MenuView int

const (
	ViewMenu MenuView = iota
	ViewSetup
	ViewResult
)

// Menu is the main menu model. It owns all key dispatch.
type Menu struct {
	keys     KeyMap
	help     help.Model
	showHelp bool
	cursor   int
	view     MenuView
	setup    primitives.Setup
	result   string
	width    int
	height   int
}

// NewMenu creates the main menu model.
func NewMenu() Menu {
	h := help.New()
	h.Styles.ShortDesc = HelpDescStyle
	h.Styles.ShortKey = HelpKeyStyle
	h.Styles.FullDesc = HelpDescStyle
	h.Styles.FullKey = HelpKeyStyle

	return Menu{
		keys:   DefaultKeys(),
		help:   h,
		view:   ViewMenu,
		width:  80,
		height: 24,
	}
}

// Init initializes the menu.
func (m Menu) Init() tea.Cmd {
	return nil
}

// Update handles all messages including key dispatch.
func (m Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.view == ViewSetup {
			m.setup = m.setup.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Forward non-key messages to the active sub-view
	if m.view == ViewSetup {
		var cmd tea.Cmd

		// Try spinner tick first
		m.setup, cmd = m.setup.Tick(msg)
		if cmd != nil {
			// Also forward to Update for async result messages
			var cmd2 tea.Cmd
			m.setup, cmd2 = m.setup.Update(msg)
			if m.setup.Done() {
				m.view = ViewMenu
			}
			return m, tea.Batch(cmd, cmd2)
		}

		// Forward to Update for async result messages
		m.setup, cmd = m.setup.Update(msg)
		if m.setup.Done() {
			m.view = ViewMenu
		}
		return m, cmd
	}

	return m, nil
}

func (m Menu) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if key.Matches(msg, m.keys.Quit) {
		if m.view == ViewSetup {
			m.view = ViewMenu
			return m, nil
		}
		if m.view == ViewResult {
			m.view = ViewMenu
			return m, nil
		}
		return m, tea.Quit
	}

	switch m.view {
	case ViewMenu:
		return m.handleMenuKey(msg)
	case ViewSetup:
		return m.handleSetupKey(msg)
	case ViewResult:
		if key.Matches(msg, m.keys.Enter) || key.Matches(msg, m.keys.Back) {
			m.view = ViewMenu
		}
		return m, nil
	}

	return m, nil
}

func (m Menu) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(menuItems)-1 {
			m.cursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		return m.activateMenuItem()
	}

	return m, nil
}

func (m Menu) activateMenuItem() (tea.Model, tea.Cmd) {
	item := menuItems[m.cursor]
	switch item.action {
	case ActionSetup:
		m.setup = primitives.NewSetup().SetSize(m.width, m.height)
		m.view = ViewSetup
		return m, m.setup.Init()

	case ActionDetect:
		return m.runDetect()
	case ActionList:
		return m.runList()
	case ActionSuggest:
		return m.runSuggest()
	case ActionAuto:
		return m.runAuto()
	case ActionEdit:
		return m.runEdit()
	}
	return m, nil
}

func (m Menu) handleSetupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	state := m.setup.State()

	// Manual input mode: form gets special handling
	if state == primitives.SetupManualInput {
		switch msg.String() {
		case "esc":
			m.setup = m.setup.Back()
			return m, nil
		case "tab", "down":
			m.setup = m.setup.FormNextField()
			return m, nil
		case "shift+tab", "up":
			m.setup = m.setup.FormPrevField()
			return m, nil
		case "enter":
			var cmd tea.Cmd
			m.setup, cmd = m.setup.Submit()
			return m, cmd
		default:
			var cmd tea.Cmd
			m.setup, cmd = m.setup.UpdateInput(msg)
			return m, cmd
		}
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		m.setup = m.setup.CursorUp()
	case key.Matches(msg, m.keys.Down):
		m.setup = m.setup.CursorDown()
	case key.Matches(msg, m.keys.Select):
		m.setup = m.setup.Toggle()
	case key.Matches(msg, m.keys.Back):
		m.setup = m.setup.Back()
	case key.Matches(msg, m.keys.Manual):
		m.setup = m.setup.EnterManualInput()
	case key.Matches(msg, m.keys.Enter):
		var cmd tea.Cmd
		m.setup, cmd = m.setup.Submit()
		return m, cmd
	}

	return m, nil
}

// --- Inline command runners ---

func (m Menu) runDetect() (tea.Model, tea.Cmd) {
	detector := mycelia.NewDetector(".")
	detected, err := detector.DetectPackages()
	if err != nil {
		m.result = fmt.Sprintf("Error: %v", err)
		m.view = ViewResult
		return m, nil
	}
	if len(detected) == 0 {
		m.result = "No package managers or build systems detected."
		m.view = ViewResult
		return m, nil
	}
	var sb strings.Builder
	sb.WriteString("Detected package managers and build systems:\n\n")
	for _, pkg := range detected {
		sb.WriteString(fmt.Sprintf("  %s %s (%s)\n", IconBullet, pkg.Type.Description, pkg.Path))
	}
	m.result = sb.String()
	m.view = ViewResult
	return m, nil
}

func (m Menu) runList() (tea.Model, tea.Cmd) {
	config, err := mycelia.LoadConfig()
	if err != nil {
		m.result = fmt.Sprintf("Error: %v", err)
		m.view = ViewResult
		return m, nil
	}
	var sb strings.Builder
	for _, category := range []string{"post-pull", "post-checkout"} {
		commands, err := config.GetCommands(category)
		if err != nil {
			continue
		}
		label := strings.ReplaceAll(category, "-", " ")
		sb.WriteString(fmt.Sprintf("\n%s%s commands:\n", strings.ToUpper(label[:1]), label[1:]))
		if len(commands) == 0 {
			sb.WriteString("  (none)\n")
		} else {
			for i, cmd := range commands {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, cmd.Description))
				sb.WriteString(fmt.Sprintf("     Command: %s\n", cmd.Command))
			}
		}
	}
	m.result = sb.String()
	m.view = ViewResult
	return m, nil
}

func (m Menu) runSuggest() (tea.Model, tea.Cmd) {
	detector := mycelia.NewDetector(".")
	var sb strings.Builder
	for _, category := range []string{"post-pull", "post-checkout"} {
		suggestions, err := detector.GetSuggestedCommands(category)
		if err != nil || len(suggestions) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\nSuggested %s commands:\n", category))
		for i, s := range suggestions {
			sb.WriteString(fmt.Sprintf("  %d. %s\n     %s\n", i+1, s.Description, s.Command))
		}
	}
	if sb.Len() == 0 {
		m.result = "No suggestions available."
	} else {
		m.result = sb.String()
	}
	m.view = ViewResult
	return m, nil
}

func (m Menu) runAuto() (tea.Model, tea.Cmd) {
	detector := mycelia.NewDetector(".")
	detected, err := detector.DetectPackages()
	if err != nil {
		m.result = fmt.Sprintf("Error: %v", err)
		m.view = ViewResult
		return m, nil
	}
	config, err := mycelia.LoadConfig()
	if err != nil {
		m.result = fmt.Sprintf("Error: %v", err)
		m.view = ViewResult
		return m, nil
	}

	var sb strings.Builder
	total := 0
	for _, category := range []string{"post-pull", "post-checkout"} {
		for _, pkg := range detected {
			commands, exists := pkg.Type.Commands[category]
			if !exists {
				continue
			}
			var triggerFiles []string
			if pkg.Type.DetectFile != "" {
				triggerFiles = append(triggerFiles, pkg.Type.DetectFile)
			}
			triggerFiles = append(triggerFiles, pkg.Type.DetectFiles...)

			for _, cmd := range commands {
				_ = config.AddCommand(category, mycelia.Command{
					Command:      cmd.Command,
					WorkingDir:   cmd.WorkingDir,
					Description:  cmd.Description,
					TriggerFiles: triggerFiles,
				})
				total++
				sb.WriteString(fmt.Sprintf("  %s %s (%s)\n", IconCheck, cmd.Description, category))
			}
		}
	}

	if total > 0 {
		if err := config.Save(); err != nil {
			m.result = fmt.Sprintf("Error saving: %v", err)
		} else {
			m.result = fmt.Sprintf("Added %d commands:\n\n%s", total, sb.String())
		}
	} else {
		m.result = "No commands to add."
	}
	m.view = ViewResult
	return m, nil
}

func (m Menu) runEdit() (tea.Model, tea.Cmd) {
	if err := mycelia.OpenConfigInEditor(); err != nil {
		m.result = fmt.Sprintf("Error: %v", err)
		m.view = ViewResult
		return m, nil
	}
	m.result = "Editor closed."
	m.view = ViewResult
	return m, nil
}

// View renders the current screen.
func (m Menu) View() string {
	switch m.view {
	case ViewSetup:
		return m.setup.View()
	case ViewResult:
		return m.viewResult()
	default:
		return m.viewMenu()
	}
}

func (m Menu) viewMenu() string {
	banner := lipgloss.NewStyle().
		Foreground(ColorTitle).
		Bold(true).
		Render(MyceliaASCII)

	var options []string
	for i, item := range menuItems {
		cursor := "  "
		if m.cursor == i {
			cursor = IconCursor + " "
		}

		line := fmt.Sprintf("%s%s %s", cursor, item.icon, item.label)
		desc := "    " + item.desc

		if m.cursor == i {
			line = SelectedItemStyle.Render(line)
			desc = SubtleTextStyle.Render(desc)
		} else {
			line = ItemStyle.Render(line)
			desc = HelpDescStyle.Render(desc)
		}
		options = append(options, line, desc)
		if i < len(menuItems)-1 {
			options = append(options, "")
		}
	}

	menuBox := BoxStyle.Width(60).Render(
		lipgloss.JoinVertical(lipgloss.Left, options...),
	)

	m.help.ShowAll = m.showHelp
	helpView := HelpStyle.Render(m.help.View(m.keys))

	content := lipgloss.JoinVertical(lipgloss.Center, banner, menuBox, helpView)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Menu) viewResult() string {
	title := TitleStyle.Render("RESULT")
	w := 70
	if m.width > 20 {
		w = min(m.width-10, 70)
	}
	box := BoxStyle.Width(w).Render(TextStyle.Render(m.result))
	helpText := HelpDescStyle.Margin(ComponentGap, 0, 0, 0).Render("enter/esc return to menu")
	content := lipgloss.JoinVertical(lipgloss.Center, title, "", box, helpText)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
