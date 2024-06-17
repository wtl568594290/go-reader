package views

import (
	"fmt"
	"go-reader/dao"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type itemShelf struct {
	title, desc string
}

type modelShelf struct {
	Selected  itemShelf
	Importing bool
	list      list.Model
}
type keyMapShelf struct {
	Select key.Binding
	Import key.Binding
	Remove key.Binding
}
type shelfMsg struct {
	msg string
}

func shelfCmd(msg shelfMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

var _docStyle = lipgloss.NewStyle().Margin(1, 2)

var _keysShelf = keyMapShelf{
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Import: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "import"),
	),
	Remove: key.NewBinding(
		key.WithKeys("r", "delete"), // "delete" is an alias for "r
		key.WithHelp("r", "remove"),
	),
}

func (i itemShelf) Title() string       { return i.title }
func (i itemShelf) Description() string { return i.desc }
func (i itemShelf) FilterValue() string { return i.title }

func (m modelShelf) Init() tea.Cmd {
	return nil
}

func (m modelShelf) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		itemLen := len(m.list.Items())
		if itemLen > 0 {
			switch {
			case key.Matches(msg, _keysShelf.Select):
				m.Selected = m.list.Items()[m.list.Index()].(itemShelf)
				pos, err := ProcBook(m.Selected.title)
				if err != nil {
					return m, dialogCmd(dialogMsg{
						Type:    DialogAlert,
						Title:   "Open " + m.Selected.title + " failed",
						Confirm: "OK",
					})
				}
				title, content, index := GetBookContent(pos)
				return m, tea.Batch(
					pagerCmd(pagerMsg{title: title, content: content, lastPos: pos, currentIndex: index}),
					viewCmd(viewPager),
				)
			case key.Matches(msg, _keysShelf.Remove):
				m.Selected = m.list.Items()[m.list.Index()].(itemShelf)
				return m, dialogCmd(dialogMsg{
					Type:    DialogDefault,
					Title:   "Delete " + m.Selected.title + "?",
					Confirm: "Delete",
					Cancel:  "Cancel",
					ConfirmFunc: func() tea.Cmd {
						err := DelBook(m.Selected.title)
						if err != nil {
							return dialogCmd(dialogMsg{
								Type:    DialogAlert,
								Title:   "Delete Failed",
								Confirm: "OK",
							})
						}
						return tea.Batch(dialogCmd(dialogMsg{Type: DialogNone}), shelfCmd(shelfMsg{msg: "refresh"}))
					},
				})
			}
		}
		if key.Matches(msg, _keysShelf.Import) {
			return m, viewCmd(viewImport)
		}
	case tea.WindowSizeMsg:
		h, v := _docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case shelfMsg:
		switch msg.msg {
		case "refresh":
			cmd := m.list.SetItems(getLatestItems())
			m.list.ResetSelected()
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelShelf) View() string {
	return _docStyle.Render(m.list.View())
}
func getLatestItems() []list.Item {

	itemShelfs := []list.Item{}
	books := dao.GetBooks()
	for _, book := range books {
		progress := book.LastPos * 100 / (book.Length - 1)
		if progress > 100 {
			progress = 100
		}
		itemShelfs = append(itemShelfs, itemShelf{title: book.Title, desc: fmt.Sprintf("进度:%d%%", progress)})
	}
	return itemShelfs
}

func NewShelf() modelShelf {
	myList := list.New(getLatestItems(), list.NewDefaultDelegate(), 0, 0)
	myList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{_keysShelf.Select, _keysShelf.Import, _keysShelf.Remove}
	}

	myList.Title = "Book Shelf"
	myList.Styles.Title = titleStyle
	m := modelShelf{
		list: myList,
	}

	return m
}
