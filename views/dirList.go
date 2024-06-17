package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DirMsg struct {
	index int
}

func dirCmd(msg DirMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type keyMapDir struct {
	Back   key.Binding
	Select key.Binding
}

var (
	// listTitleStyle    = lipgloss.NewStyle().MarginLeft(2)
	_itemDirStyle      = lipgloss.NewStyle().PaddingLeft(4)
	_selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	_keysDir           = keyMapDir{
		Back: key.NewBinding(
			key.WithKeys("q", "esc", "d"),
			key.WithHelp("d", "Back"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Select"),
		)}
)

type itemDir BookDir

func (i itemDir) FilterValue() string { return i.name }

type itemDirDelegate struct{}

func (d itemDirDelegate) Height() int                             { return 1 }
func (d itemDirDelegate) Spacing() int                            { return 0 }
func (d itemDirDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDirDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(itemDir)
	if !ok {
		return
	}

	str := fmt.Sprintf(" %s", i.name)

	fn := _itemDirStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return _selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type modelList struct {
	list   list.Model
	choice itemDir
}

func (m modelList) Init() tea.Cmd {
	return nil
}

func (m modelList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, _keysDir.Back):
			return m, viewCmd(viewMsg(viewPager))

		case key.Matches(msg, _keysDir.Select):
			i, ok := m.list.SelectedItem().(itemDir)
			if ok {
				m.choice = i
			}
			if m.list.Index() >= 0 {
				title, content, index := GetBookContent(bookDirs[m.list.Index()].start)
				return m, tea.Batch(
					pagerCmd(pagerMsg{title: title, content: content, lastPos: GetChapterStart(m.list.Index()), currentIndex: index}),
					viewCmd(viewPager),
				)
			}
			return m, nil
		}

	case DirMsg:
		items := []list.Item{}
		for _, dir := range bookDirs {
			items = append(items, itemDir(dir))
		}
		cmds = append(cmds, m.list.SetItems(items))
		m.list.Select(msg.index)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m modelList) View() string {
	return "\n" + m.list.View()
}

func NewDirList() modelList {
	itemDirs := []list.Item{}
	l := list.New(itemDirs, itemDirDelegate{}, 20, 20)
	l.Title = "Directory List"

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{_keysDir.Select}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{_keysDir.Select}
	}
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	m := modelList{list: l}

	return m
}
