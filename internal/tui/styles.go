package tui

import "github.com/charmbracelet/lipgloss"

// Spacing constants
const (
	SpacingUnit    = 1
	DefaultMargin  = 2
	DefaultPadding = 1
	ListIndent     = 2
	ComponentGap   = 1
	SectionGap     = 2
)

// Color palette — Tokyo Night
var (
	ColorTitle     = lipgloss.Color("#7DCFFF")
	ColorTitleAlt  = lipgloss.Color("#2AC3DE")
	ColorAccent    = lipgloss.Color("#BB9AF7")
	ColorAccentAlt = lipgloss.Color("#9D7CD8")

	ColorSuccess    = lipgloss.Color("#9ECE6A")
	ColorSuccessAlt = lipgloss.Color("#73DACA")
	ColorWarning    = lipgloss.Color("#E0AF68")
	ColorWarningAlt = lipgloss.Color("#FF9E64")
	ColorError      = lipgloss.Color("#F7768E")
	ColorErrorAlt   = lipgloss.Color("#DB4B4B")
	ColorInfo       = lipgloss.Color("#7AA2F7")

	ColorPrimary   = lipgloss.Color("#C0CAF5")
	ColorSecondary = lipgloss.Color("#565F89")
	ColorTertiary  = lipgloss.Color("#414868")
	ColorSubtle    = lipgloss.Color("#3B4261")
	ColorMuted     = lipgloss.Color("#545c7e")

	ColorHighlight    = lipgloss.Color("#FF9E64")
	ColorSelected     = lipgloss.Color("#ff9e64")
	ColorBorder       = lipgloss.Color("#7AA2F7")
	ColorBorderDim    = lipgloss.Color("#3D59A1")
	ColorBorderAccent = lipgloss.Color("#BB9AF7")

	ColorBase        = lipgloss.Color("#1a1b26")
	ColorBaseLighter = lipgloss.Color("#24283b")
	ColorOverlay     = lipgloss.Color("#292e42")
)

// Icons
const (
	IconCheck    = "✓"
	IconCross    = "×"
	IconWarning  = "⚠"
	IconInfo     = "ⓘ"
	IconSettings = "⚙"
	IconCursor   = "❯"
	IconBullet   = "•"
	IconCheckbox = "☐"
	IconChecked  = "☑"
	IconArrow    = "→"
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorTitle).
			Bold(true).
			Padding(0, DefaultPadding).
			Margin(DefaultMargin, 0)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, DefaultPadding)

	TextStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	SubtleTextStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Padding(DefaultPadding, 0)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorTertiary)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	ItemStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			PaddingLeft(ListIndent)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Bold(true).
				PaddingLeft(ListIndent)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(DefaultPadding, DefaultPadding*2)

	ActiveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ColorTitle).
			Padding(DefaultPadding, DefaultPadding*2)

	DimBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorderDim).
			Padding(DefaultPadding, DefaultPadding*2)
)

// MyceliaASCII is the ASCII art banner for Mycelia.
const MyceliaASCII = `
  __  __                _ _
 |  \/  |_   _  ___ ___| (_) __ _
 | |\/| | | | |/ __/ _ \ | |/ _` + "`" + ` |
 | |  | | |_| | (_|  __/ | | (_| |
 |_|  |_|\__, |\___\___|_|_|\__,_|
         |___/
`
