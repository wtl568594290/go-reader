package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
)

func DialogBox(title string, confirm string, cancel string, width int) string {
	if width == 0 {
		width = 96
	}
	okButton := activeButtonStyle.Render(confirm)
	cancelButton := buttonStyle.Render(cancel)

	question := lipgloss.NewStyle().Width(50).Align(lipgloss.Center).Render(title)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
	ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)

	dialog := lipgloss.Place(width, 9,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#874BFD")),
		lipgloss.WithWhitespaceForeground(subtle),
	)

	return dialog
}

func Alert(title string, confirm string, width int) string {
	if width == 0 {
		width = 96
	}
	okButton := activeButtonStyle.Render(confirm)
	question := lipgloss.NewStyle().Width(50).Align(lipgloss.Center).Render(title)
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, okButton)
	ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)
	dialog := lipgloss.Place(width, 9,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#874BFD")),
		lipgloss.WithWhitespaceForeground(subtle),
	)
	return dialog

}

func AppendDialog(src string, dialog string, height int) string {
	h := lipgloss.Height(dialog)
	if h < height {
		start := (height - h) / 2
		// 替换中间行为dialog
		lines := strings.Split(src, "\n")
		dialogLines := strings.Split(dialog, "\n")
		for i := start; i < start+h; i++ {
			lines[i] = dialogLines[i-start]
		}
		src = strings.Join(lines, "\n")
	}
	return src
}
