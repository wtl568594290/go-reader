package views

import (
	"os"

	"go-reader/utils"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type keyMapImport struct {
	Up         key.Binding
	Down       key.Binding
	Back       key.Binding
	Open       key.Binding
	Top        key.Binding
	Bottom     key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Select     key.Binding
	Help       key.Binding
	Quit       key.Binding
	SwitchDisk map[string]key.Binding
}

type modelImport struct {
	keysImport   keyMapImport
	help         help.Model
	filepicker   filepicker.Model
	selectedFile string
}

// full help 行数-1
// 以help中key.binding中的行数为准
const (
	_fullThanShort = 3
	_marginBottom  = 6
)

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMapImport) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Select, k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMapImport) FullHelp() [][]key.Binding {
	fullHelp := key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "short help"))
	switchDisk := key.NewBinding(key.WithKeys("alt+<disk>"), key.WithHelp("alt+<disk>", "switch disk"))
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Back, k.Open, k.Select, k.Quit},
		{fullHelp, switchDisk},
	}
}

func (m modelImport) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m modelImport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		winheight = msg.Height
		m.filepicker.Height = winheight - _marginBottom
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keysImport.Help):
			m.help.ShowAll = !m.help.ShowAll
			var height int
			if m.help.ShowAll {
				height = winheight - _marginBottom - _fullThanShort
			} else {
				height = winheight - _marginBottom
			}
			m.filepicker.Height = height
			// 重设高度
			msgSetHeight := tea.WindowSizeMsg{Width: 0, Height: height}
			m.filepicker, _ = m.filepicker.Update(msgSetHeight)
			// 调用一次Home键，将光标移动到第一行
			return m, keyCmd(tea.KeyMsg{Type: tea.KeyHome})
		case key.Matches(msg, m.keysImport.Quit):
			return m, viewCmd(viewShelf)
		case key.Matches(msg, m.keysImport.Down):
			msgDown := tea.KeyMsg{Type: tea.KeyDown}
			m.filepicker, _ = m.filepicker.Update(msgDown)
			return m, nil

		/** 原生filepicker翻页有BUG,屏蔽它 **/
		case key.Matches(msg, m.keysImport.PageUp, m.keysImport.PageDown):
			return m, nil
		default:
			for disk, binding := range m.keysImport.SwitchDisk {
				if key.Matches(msg, binding) {
					driverLetter := disk + ":\\"
					if utils.HasDisk(driverLetter) {
						m.filepicker.CurrentDirectory = disk + ":\\"
						return m, m.filepicker.Init()
					} else {
						return m, dialogCmd(dialogMsg{Type: DialogAlert, Title: "open " + driverLetter + " failed", Confirm: "OK"})
					}
				}
			}
		}
	}

	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
		// return m, tea.Quit
		err := ImportBook(path)
		if err != nil {
			// m.err = errors.New("Import failed for " + path + ".")
			m.selectedFile = ""
			return m, tea.Batch(cmd, dialogCmd(dialogMsg{Type: DialogAlert, Title: err.Error(), Confirm: "OK"}))
		}

		return m, tea.Batch(cmd, viewCmd(viewShelf), shelfCmd(shelfMsg{msg: "refresh"}))
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, _ := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.selectedFile = ""
		return m, tea.Batch(cmd, dialogCmd(dialogMsg{Type: DialogAlert, Title: "Unsupported file", Confirm: "OK"}))
	}

	return m, cmd
}

func (m modelImport) View() string {
	s := "\n"
	s += titleStyle.Render("Import Book")
	s += subTitleStyle.Render(m.filepicker.CurrentDirectory)
	s += "\n\n" + m.filepicker.View() + "\n"
	helpView := m.help.View(m.keysImport)
	s += helpView
	return s
}

func NewImport() modelImport {
	fp := filepicker.New()
	fp.AutoHeight = false
	fp.AllowedTypes = []string{".txt"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	switchDisk := make(map[string]key.Binding)
	for i := 'C'; i <= 'Z'; i++ {
		// 转为小写字母
		lowcase := string(i + 32)
		switchDisk[string(i)] = key.NewBinding(
			key.WithKeys("alt+" + lowcase))
	}

	keysImport := keyMapImport{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Back: key.NewBinding(
			key.WithKeys("left", "h", "backspace"),
			key.WithHelp("←/h/backspace", "back"),
		),
		Open: key.NewBinding(
			key.WithKeys("right", "l", "enter"),
			key.WithHelp("→/l/enter", "open"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"), // "G" is shift + "g
			key.WithHelp("end/G", "bottom"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "J"),
			key.WithHelp("pgup/J", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "K"),
			key.WithHelp("pgdown/K", "page down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "quit"),
		),
		SwitchDisk: switchDisk,
	}

	m := modelImport{
		filepicker: fp,
		keysImport: keysImport,
		help:       help.New(),
	}
	return m
}
