package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7C3AED")).
			PaddingLeft(1).
			PaddingRight(1)
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1A1AA")).
			Italic(true)
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E")).
			Background(lipgloss.Color("#052E16")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
	KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true)
	RootStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)
	UserStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)
	KillStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)
	ProcessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F97316")).
			Bold(true)
)
