package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMsg int

func viewCmd(v viewMsg) tea.Cmd {
	return func() tea.Msg {
		return v
	}
}
func keyCmd(key tea.KeyMsg) tea.Cmd {
	return func() tea.Msg {
		return key
	}
}

type modelViews struct {
	models []tea.Model
	dialog dialog
}

type keyMapViews struct {
	ForceQuit key.Binding // 强制退出
}

const (
	viewShelf = iota
	viewImport
	viewPager
	viewDirList
)

var winwidth, winheight int
var titleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
var subTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Padding(0, 2)

var _winTitle = []string{"BookShelf", "Import Book", "Reading", "Directory List"}
var _curView = viewShelf
var _keysViews = keyMapViews{
	ForceQuit: key.NewBinding(key.WithKeys("x", "ctrl+c")),
}

func (v modelViews) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, v.dialog.Init())
	for _, model := range v.models {
		cmds = append(cmds, model.Init())
	}
	cmds = append(cmds, viewCmd(viewShelf))
	return tea.Batch(cmds...)
}

func (v modelViews) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, _keysViews.ForceQuit) {
			return v, tea.Quit
		}
		// tea.KeyMsg 只执行当前step的update,当有dialog时只执行dialog的update
		if v.dialog.Type != DialogNone {
			v.dialog, cmd = v.dialog.Update(msg)
		} else {
			v.models[_curView], cmd = v.models[_curView].Update(msg)
		}
		return v, cmd
	case tea.WindowSizeMsg:
		winwidth, winheight = msg.Width, msg.Height
	case viewMsg:
		_curView = int(msg)
		cmds = append(cmds, tea.SetWindowTitle(_winTitle[_curView]))
	}

	v.dialog, cmd = v.dialog.Update(msg)
	cmds = append(cmds, cmd)
	for i, model := range v.models {
		v.models[i], cmd = model.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v modelViews) View() string {
	// return v.models[step].View()
	return v.dialog.AppendDialog(v.models[_curView].View())
}

func NewViews() {
	dialog := NewDialog()
	shelf := NewShelf()
	imp := NewImport()
	pager := NewPager()
	dirList := NewDirList()

	models := []tea.Model{shelf, imp, pager, dirList}
	m := modelViews{
		models: models,
		dialog: dialog,
	}
	tea.NewProgram(m, tea.WithAltScreen()).Run()
}
