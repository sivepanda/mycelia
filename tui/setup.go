package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sivepanda/mycelia"
)

// SetupState represents the current phase of the setup wizard.
type SetupState int

const (
	SetupDetecting SetupState = iota
	SetupPackageSelect
	SetupCategorySelect
	SetupCommandSelect
	SetupManualInput
	SetupConfirm
	SetupExecute
	SetupComplete
)

// SetupStyles configures the visual appearance of the Setup wizard.
type SetupStyles struct {
	Title       lipgloss.Style
	Header      lipgloss.Style
	Text        lipgloss.Style
	Subtle      lipgloss.Style
	Help        lipgloss.Style
	Success     lipgloss.Style
	Error       lipgloss.Style
	Box         lipgloss.Style
	ActiveBox   lipgloss.Style
	DimBox      lipgloss.Style
	ItemStyle   lipgloss.Style
	ActiveItem  lipgloss.Style
	BulletIcon  string
	SpinnerColor lipgloss.Color
}

// DefaultSetupStyles returns Tokyo Night themed styles.
func DefaultSetupStyles() SetupStyles {
	return SetupStyles{
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7DCFFF")).
			Bold(true).
			Padding(0, 1).
			Margin(2, 0),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB9AF7")).
			Bold(true).
			Padding(0, 1),
		Text: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C0CAF5")),
		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89")),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#414868")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ECE6A")).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F7768E")).
			Bold(true),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7AA2F7")).
			Padding(1, 2),
		ActiveBox: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#7DCFFF")).
			Padding(1, 2),
		DimBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#3D59A1")).
			Padding(1, 2),
		ItemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C0CAF5")),
		ActiveItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9E64")).
			Bold(true),
		BulletIcon:   "•",
		SpinnerColor: lipgloss.Color("#BB9AF7"),
	}
}

// SuggestionItem represents a command suggestion with selection state.
type SuggestionItem struct {
	Command      mycelia.Command
	TriggerFiles []string
	Selected     bool
}

// CategoryItem represents a category with selection state.
type CategoryItem struct {
	Name     string
	Selected bool
}

// PackageItem represents a detected package with selection state.
type PackageItem struct {
	Package  mycelia.DetectedPackage
	Selected bool
}

// DetectionCompleteMsg indicates package detection is complete.
type DetectionCompleteMsg struct {
	Detected []mycelia.DetectedPackage
	Config   *mycelia.Config
	Error    error
}

// SuggestionsLoadedMsg indicates suggestions have been loaded.
type SuggestionsLoadedMsg struct {
	Category    string
	Suggestions []SuggestionItem
	Error       error
}

// CommandsAddedMsg indicates commands have been added to config.
type CommandsAddedMsg struct {
	Count    int
	Category string
	Error    error
}

// Setup is a headless setup wizard primitive.
// It never handles tea.KeyMsg — the parent drives it via imperative methods.
// Update() only processes internal async messages (DetectionCompleteMsg, etc.).
type Setup struct {
	state           SetupState
	spinner         spinner.Model
	detector        *mycelia.Detector
	detected        []mycelia.DetectedPackage
	config          *mycelia.Config
	packages        Selector[PackageItem]
	categories      Selector[CategoryItem]
	commands        Selector[SuggestionItem]
	form            Form
	currentCategory int
	addedCount      int
	err             error
	width           int
	height          int
	hPos            lipgloss.Position
	vPos            lipgloss.Position
	styles          SetupStyles
}

// NewSetup creates a new Setup wizard primitive.
func NewSetup() Setup {
	styles := DefaultSetupStyles()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.SpinnerColor)

	categories := NewSelector([]CategoryItem{
		{Name: "post-pull", Selected: true},
		{Name: "post-checkout", Selected: true},
	}, func(c CategoryItem) string { return c.Name }).
		WithSelected(true)

	form := NewForm([]FormField{
		{Key: "command", Label: "Command:", Placeholder: "e.g., npm run build"},
		{Key: "workingDir", Label: "Working Directory:", Placeholder: "e.g., ."},
		{Key: "description", Label: "Description:", Placeholder: "e.g., Build the project"},
	})

	return Setup{
		state:      SetupDetecting,
		spinner:    s,
		detector:   mycelia.NewDetector("."),
		categories: categories,
		form:       form,
		width:      80,
		styles:     styles,
	}
}

// WithStyles replaces the style config.
func (s Setup) WithStyles(styles SetupStyles) Setup {
	s.styles = styles
	s.spinner.Style = lipgloss.NewStyle().Foreground(styles.SpinnerColor)
	return s
}

// WithPosition sets the horizontal and vertical placement within the
// available width/height. Use lipgloss.Center for both to center the wizard.
func (s Setup) WithPosition(hPos, vPos lipgloss.Position) Setup {
	s.hPos = hPos
	s.vPos = vPos
	return s
}

// --- Init command (parent should call this and feed the resulting Cmd) ---

// Init returns the initial commands: spinner tick + package detection.
func (s Setup) Init() tea.Cmd {
	return tea.Batch(s.spinner.Tick, s.detectPackages())
}

// --- Internal async commands ---

func (s Setup) detectPackages() tea.Cmd {
	return func() tea.Msg {
		detected, err := s.detector.DetectPackages()
		if err != nil {
			return DetectionCompleteMsg{Error: err}
		}
		config, err := mycelia.LoadConfig()
		if err != nil {
			return DetectionCompleteMsg{Error: err}
		}
		return DetectionCompleteMsg{
			Detected: detected,
			Config:   config,
		}
	}
}

func (s Setup) getSuggestions() tea.Cmd {
	return func() tea.Msg {
		selectedCats := s.categories.Selected()
		if s.currentCategory >= len(selectedCats) {
			return SuggestionsLoadedMsg{Error: fmt.Errorf("no more categories")}
		}
		categoryName := selectedCats[s.currentCategory].Name

		selectedPkgs := s.packages.Selected()
		var items []SuggestionItem
		for _, pkgItem := range selectedPkgs {
			for _, pkgType := range mycelia.PackageTypes {
				if pkgType.Name != pkgItem.Package.Type.Name {
					continue
				}
				var triggerFiles []string
				if pkgType.DetectFile != "" {
					triggerFiles = append(triggerFiles, pkgType.DetectFile)
				}
				triggerFiles = append(triggerFiles, pkgType.DetectFiles...)

				if commands, exists := pkgType.Commands[categoryName]; exists {
					for _, cmd := range commands {
						items = append(items, SuggestionItem{
							Command:      cmd,
							TriggerFiles: triggerFiles,
							Selected:     true,
						})
					}
				}
				break
			}
		}

		return SuggestionsLoadedMsg{
			Category:    categoryName,
			Suggestions: items,
		}
	}
}

func (s Setup) addSelectedCommands() tea.Cmd {
	return func() tea.Msg {
		selectedCats := s.categories.Selected()
		if s.currentCategory >= len(selectedCats) {
			return CommandsAddedMsg{Error: fmt.Errorf("no category")}
		}
		categoryName := selectedCats[s.currentCategory].Name

		count := 0
		for _, item := range s.commands.Items() {
			if !item.Selected {
				continue
			}
			cmd := mycelia.Command{
				Command:      item.Value.Command.Command,
				WorkingDir:   item.Value.Command.WorkingDir,
				Description:  item.Value.Command.Description,
				TriggerFiles: item.Value.TriggerFiles,
			}
			if err := s.config.AddCommand(categoryName, cmd); err != nil {
				return CommandsAddedMsg{Error: err}
			}
			count++
		}

		if count > 0 {
			if err := s.config.Save(); err != nil {
				return CommandsAddedMsg{Error: err}
			}
		}

		return CommandsAddedMsg{
			Count:    count,
			Category: categoryName,
		}
	}
}

// --- Update: only handles internal async messages ---

// Update processes internal messages (detection results, spinner ticks, etc.).
// It does NOT handle tea.KeyMsg — the parent must use the imperative methods instead.
func (s Setup) Update(msg tea.Msg) (Setup, tea.Cmd) {
	switch msg := msg.(type) {
	case DetectionCompleteMsg:
		if msg.Error != nil {
			s.err = msg.Error
			s.state = SetupComplete
			return s, nil
		}
		s.detected = msg.Detected
		s.config = msg.Config

		if len(s.detected) == 0 {
			s.err = fmt.Errorf("no package managers detected")
			s.state = SetupComplete
			return s, nil
		}

		pkgItems := make([]PackageItem, len(s.detected))
		for i, pkg := range s.detected {
			pkgItems[i] = PackageItem{Package: pkg, Selected: true}
		}
		s.packages = NewSelector(pkgItems, func(p PackageItem) string {
			return p.Package.Type.Description
		}).WithSelected(true)

		s.state = SetupPackageSelect
		return s, nil

	case SuggestionsLoadedMsg:
		if msg.Error != nil {
			s.err = msg.Error
			s.state = SetupComplete
			return s, nil
		}
		if len(msg.Suggestions) == 0 {
			s.err = fmt.Errorf("no suggestions for %s", msg.Category)
			s.state = SetupComplete
			return s, nil
		}

		s.commands = NewSelector(msg.Suggestions, func(si SuggestionItem) string {
			return si.Command.Description
		}).WithDetail(func(si SuggestionItem) string {
			return si.Command.Command
		}).WithSelected(true)

		s.state = SetupCommandSelect
		return s, nil

	case CommandsAddedMsg:
		if msg.Error != nil {
			s.err = msg.Error
			s.state = SetupComplete
			return s, nil
		}
		s.addedCount += msg.Count

		s.currentCategory++
		selectedCats := s.categories.Selected()
		if s.currentCategory < len(selectedCats) {
			return s, s.getSuggestions()
		}

		s.state = SetupComplete
		return s, nil

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	}

	return s, nil
}

// --- Imperative control methods ---

// CursorUp moves the cursor up in the current state's list.
func (s Setup) CursorUp() Setup {
	switch s.state {
	case SetupPackageSelect:
		s.packages = s.packages.CursorUp()
	case SetupCategorySelect:
		s.categories = s.categories.CursorUp()
	case SetupCommandSelect:
		s.commands = s.commands.CursorUp()
	}
	return s
}

// CursorDown moves the cursor down in the current state's list.
func (s Setup) CursorDown() Setup {
	switch s.state {
	case SetupPackageSelect:
		s.packages = s.packages.CursorDown()
	case SetupCategorySelect:
		s.categories = s.categories.CursorDown()
	case SetupCommandSelect:
		s.commands = s.commands.CursorDown()
	}
	return s
}

// Toggle flips the selection of the item under the cursor.
func (s Setup) Toggle() Setup {
	switch s.state {
	case SetupPackageSelect:
		s.packages = s.packages.Toggle()
	case SetupCategorySelect:
		s.categories = s.categories.Toggle()
	case SetupCommandSelect:
		s.commands = s.commands.Toggle()
	}
	return s
}

// Submit advances to the next state. Returns a Cmd when async work is needed.
func (s Setup) Submit() (Setup, tea.Cmd) {
	switch s.state {
	case SetupPackageSelect:
		if !s.packages.HasSelected() {
			s.err = fmt.Errorf("no packages selected")
			s.state = SetupComplete
			return s, nil
		}
		s.state = SetupCategorySelect
		return s, nil

	case SetupCategorySelect:
		if !s.categories.HasSelected() {
			s.err = fmt.Errorf("no categories selected")
			s.state = SetupComplete
			return s, nil
		}
		s.currentCategory = 0
		return s, s.getSuggestions()

	case SetupCommandSelect:
		if !s.commands.HasSelected() {
			s.err = fmt.Errorf("no commands selected")
			s.state = SetupComplete
			return s, nil
		}
		s.state = SetupConfirm
		return s, nil

	case SetupManualInput:
		vals := s.form.Values()
		cmd := vals["command"]
		workingDir := vals["workingDir"]
		desc := vals["description"]

		if cmd == "" {
			s.err = fmt.Errorf("command cannot be empty")
			s.state = SetupComplete
			return s, nil
		}
		if workingDir == "" {
			workingDir = "."
		}
		if desc == "" {
			desc = cmd
		}

		// Add to the command suggestions list
		newItems := make([]SuggestionItem, s.commands.Len())
		for i, item := range s.commands.Items() {
			newItems[i] = item.Value
		}
		newItems = append(newItems, SuggestionItem{
			Command: mycelia.Command{
				Command:     cmd,
				WorkingDir:  workingDir,
				Description: desc,
			},
			Selected: true,
		})
		s.commands = NewSelector(newItems, func(si SuggestionItem) string {
			return si.Command.Description
		}).WithDetail(func(si SuggestionItem) string {
			return si.Command.Command
		}).WithSelected(true)

		s.form = s.form.Reset()
		s.state = SetupCommandSelect
		return s, nil

	case SetupConfirm:
		s.state = SetupExecute
		return s, tea.Batch(s.spinner.Tick, s.addSelectedCommands())

	case SetupComplete:
		return s, tea.Quit
	}

	return s, nil
}

// Back goes back one state.
func (s Setup) Back() Setup {
	switch s.state {
	case SetupCategorySelect:
		s.state = SetupPackageSelect
	case SetupCommandSelect:
		s.state = SetupCategorySelect
	case SetupConfirm:
		s.state = SetupCommandSelect
	case SetupManualInput:
		s.form = s.form.Reset()
		s.state = SetupCommandSelect
	}
	return s
}

// EnterManualInput switches to the manual input form (only valid during CommandSelect).
func (s Setup) EnterManualInput() Setup {
	if s.state == SetupCommandSelect {
		s.state = SetupManualInput
	}
	return s
}

// Tick forwards a message to internal spinners.
func (s Setup) Tick(msg tea.Msg) (Setup, tea.Cmd) {
	if s.state == SetupDetecting || s.state == SetupExecute {
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}
	return s, nil
}

// UpdateInput forwards a tea.Msg to the form's focused text input.
func (s Setup) UpdateInput(msg tea.Msg) (Setup, tea.Cmd) {
	if s.state == SetupManualInput {
		var cmd tea.Cmd
		s.form, cmd = s.form.UpdateInput(msg)
		return s, cmd
	}
	return s, nil
}

// FormNextField moves focus to the next form field.
func (s Setup) FormNextField() Setup {
	if s.state == SetupManualInput {
		s.form = s.form.NextField()
	}
	return s
}

// FormPrevField moves focus to the previous form field.
func (s Setup) FormPrevField() Setup {
	if s.state == SetupManualInput {
		s.form = s.form.PrevField()
	}
	return s
}

// SetSize updates the width and height.
func (s Setup) SetSize(w, h int) Setup {
	s.width = w
	s.height = h
	return s
}

// --- State inspection ---

// State returns the current setup state.
func (s Setup) State() SetupState {
	return s.state
}

// Done returns true when the wizard is complete.
func (s Setup) Done() bool {
	return s.state == SetupComplete
}

// Err returns the current error, if any.
func (s Setup) Err() error {
	return s.err
}

// AddedCount returns the number of commands added.
func (s Setup) AddedCount() int {
	return s.addedCount
}

// --- Rendering ---

// View renders the current state of the setup wizard.
func (s Setup) View() string {
	st := s.styles
	boxWidth := 60
	if s.width > 20 {
		boxWidth = min(s.width-10, 70)
	}

	var content string

	switch s.state {
	case SetupDetecting:
		title := st.Title.Render("⚙ HOUSEKEEPING SETUP")
		text := st.Text.Render("Detecting package managers and build systems...")
		box := st.Box.Width(boxWidth).Render(s.spinner.View() + " " + text)
		content = lipgloss.JoinVertical(s.hPos, title, "", box)

	case SetupPackageSelect:
		title := st.Title.Render("✓ DETECTED PACKAGES")
		header := st.Header.Margin(0, 0, 1, 0).Render("Select which package managers to use:")
		box := st.Box.Width(boxWidth).Render(s.packages.View())
		help := st.Help.Margin(1, 0, 0, 0).Render("↑/↓ navigate • x toggle • enter continue")
		content = lipgloss.JoinVertical(s.hPos, title, "", header, box, help)

	case SetupCategorySelect:
		title := st.Title.Render("✓ SELECTED PACKAGES")

		// Show selected packages summary
		var selectedList []string
		for _, item := range s.packages.Items() {
			if item.Selected {
				selectedList = append(selectedList, st.Subtle.Render("  "+st.BulletIcon)+" "+st.Text.Render(item.Value.Package.Type.Description))
			}
		}
		pkgBox := st.DimBox.Width(boxWidth).Render(strings.Join(selectedList, "\n"))

		catHeader := st.Header.Margin(2, 0, 1, 0).Render("Select categories to configure:")
		catBox := st.Box.Width(boxWidth).Render(s.categories.View())
		help := st.Help.Margin(1, 0, 0, 0).Render("↑/↓ navigate • x toggle • enter continue")
		content = lipgloss.JoinVertical(s.hPos, title, "", pkgBox, catHeader, catBox, help)

	case SetupCommandSelect:
		selectedCats := s.categories.Selected()
		catName := ""
		if s.currentCategory < len(selectedCats) {
			catName = selectedCats[s.currentCategory].Name
		}
		title := st.Title.Render(fmt.Sprintf("⚙ %s COMMANDS", strings.ToUpper(catName)))
		box := st.ActiveBox.Width(boxWidth + 10).Render(s.commands.View())
		help := st.Help.Margin(1, 0, 0, 0).Render("↑/↓ navigate • x toggle • i add manual • enter continue")
		content = lipgloss.JoinVertical(s.hPos, title, "", box, help)

	case SetupManualInput:
		selectedCats := s.categories.Selected()
		catName := ""
		if s.currentCategory < len(selectedCats) {
			catName = selectedCats[s.currentCategory].Name
		}
		title := st.Title.Render(fmt.Sprintf("⚙ ADD MANUAL COMMAND (%s)", strings.ToUpper(catName)))
		header := st.Header.Margin(0, 0, 1, 0).Render("Enter command details:")
		box := st.Box.Width(boxWidth + 10).Render(s.form.View())
		help := st.Help.Margin(1, 0, 0, 0).Render("tab/↑/↓ navigate fields • enter submit • esc cancel")
		content = lipgloss.JoinVertical(s.hPos, title, "", header, box, help)

	case SetupConfirm:
		title := st.Title.Render("✓ CONFIRM SELECTION")

		selectedCount := 0
		var selectedList []string
		for _, item := range s.commands.Items() {
			if item.Selected {
				selectedCount++
				selectedList = append(selectedList, st.Subtle.Render("  "+st.BulletIcon)+" "+st.Text.Render(item.Value.Command.Description))
			}
		}

		selectedCats := s.categories.Selected()
		catName := ""
		if s.currentCategory < len(selectedCats) {
			catName = selectedCats[s.currentCategory].Name
		}
		countHeader := st.Header.Render(fmt.Sprintf("Ready to add %d %s commands:", selectedCount, catName))
		box := st.Box.Width(boxWidth + 10).Render(strings.Join(selectedList, "\n"))
		help := st.Help.Margin(1, 0, 0, 0).Render("enter confirm • q cancel")
		content = lipgloss.JoinVertical(s.hPos, title, "", countHeader, "", box, help)

	case SetupExecute:
		title := st.Title.Render("⚙ PROCESSING")
		text := st.Text.Render("Adding selected commands to configuration...")
		box := st.Box.Width(boxWidth).Render(s.spinner.View() + " " + text)
		content = lipgloss.JoinVertical(s.hPos, title, "", box)

	case SetupComplete:
		if s.err != nil {
			title := st.Error.Render("× ERROR")
			errorMsg := st.Error.Render(fmt.Sprintf("Error: %v", s.err))
			errorBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#F7768E")).
				Padding(1, 2).
				Width(boxWidth).
				Render(errorMsg)
			help := st.Help.Margin(1, 0, 0, 0).Render("enter exit")
			content = lipgloss.JoinVertical(s.hPos, title, "", errorBox, help)
		} else {
			title := st.Success.Render("✓ COMPLETE")
			selectedCats := s.categories.Selected()
			var catNames []string
			for _, cat := range selectedCats {
				catNames = append(catNames, cat.Name)
			}
			categoryText := strings.Join(catNames, " and ")
			successMsg := st.Success.Render(fmt.Sprintf("Successfully added %d commands for %s!", s.addedCount, categoryText))
			successBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#9ECE6A")).
				Padding(1, 2).
				Width(boxWidth).
				Render(successMsg)
			help := st.Help.Margin(1, 0, 0, 0).Render("enter exit")
			content = lipgloss.JoinVertical(s.hPos, title, "", successBox, help)
		}

	default:
		return ""
	}

	if s.width > 0 && s.height > 0 {
		return lipgloss.Place(s.width, s.height, s.hPos, s.vPos, content)
	}
	return content
}
