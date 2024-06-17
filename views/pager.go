package views

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type pagerMsg struct {
	title        string
	content      string
	currentIndex int
	lastPos      int
}

func pagerCmd(pm pagerMsg) tea.Cmd {
	return func() tea.Msg {
		return pm
	}
}

type keyMapPager struct {
	PageUp   key.Binding
	PageDown key.Binding
	OpenDir  key.Binding
	Quit     key.Binding
}

var _keysPager = keyMapPager{
	PageUp: key.NewBinding(
		key.WithKeys("pageup", "left"),
		key.WithHelp("left", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pagedown", "right"),
		key.WithHelp("right", "page down"),
	),
	OpenDir: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "open dir"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("q", "quit"),
	),
}

func (k keyMapPager) ShortHelp() []key.Binding {
	return []key.Binding{k.PageUp, k.PageDown, k.OpenDir, k.Quit}
}

func (k keyMapPager) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

type modelPager struct {
	pagerMsg
	currentIndex int
	ready        bool
	help         help.Model
	viewport     viewport.Model
}

func (m modelPager) Init() tea.Cmd {
	return nil
}

var pageTotal = 1   // 总页数
var currentPage = 1 // 当前页
var jump = 0        // 跳转页数
var posMapOffset = make(map[int]int)

/** 处理文章内容，使其适应屏幕宽度和高度 */
func proc(content string, maxWidth int, maxHeight int, offset int) string {
	// pos需要+1，因为第一行是章节名
	posMapOffset[1] = 1
	maxWidth -= 2
	// 按照换行符分割字符串
	lines := strings.Split(content, "\n")
	// 将超出最大宽度的行进行分割
	newLines := make([]string, 0)
	for i, line := range lines {
		for runewidth.StringWidth(line) > maxWidth {
			cut := runewidth.Truncate(line, maxWidth, "")
			newLines = append(newLines, cut)
			line = strings.TrimPrefix(line, cut)
		}
		newLines = append(newLines, line)

		newLinesLen := len(newLines)
		if newLinesLen%maxHeight == 0 {
			posMapOffset[newLinesLen/maxHeight] = i + 1
		}
	}
	// 如果总行数不是最大高度的倍数，补充空行
	ac := len(newLines) % maxHeight
	if ac != 0 {
		for i := 0; i < maxHeight-ac; i++ {
			newLines = append(newLines, "")
		}
	}
	pageTotal = len(newLines) / maxHeight
	posMapOffset[pageTotal] = len(lines)

	currentPage = 1
	// 查看pos在第几页
	if offset > 0 && offset < len(lines) {
		jump = int(math.Floor(float64(offset)/float64(maxHeight))) + 1
	} else if offset >= len(lines) {
		jump = pageTotal
	}

	return strings.Join(newLines, "\n")
}

func (m modelPager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, _keysPager.Quit):
			cmds = append(cmds, shelfCmd(shelfMsg{msg: "refresh"}))
			cmds = append(cmds, viewCmd(viewShelf))
			return m, tea.Batch(cmds...)
		case key.Matches(msg, _keysPager.PageUp):
			if currentPage <= 1 {
				// 上一章
				if m.currentIndex >= 0 {
					title, content, index := GetBookContent(GetChapterStart(m.currentIndex - 1))
					cmds = append(cmds, pagerCmd(pagerMsg{title: title, content: content, lastPos: GetChapterStart(m.currentIndex) - 1, currentIndex: index}))
					return m, tea.Batch(cmds...)
				}
				return m, nil
			}
			currentPage--
			UpdateBookPos(bookName, GetChapterStart(m.currentIndex)+posMapOffset[currentPage])
			nextmsg := tea.KeyMsg{Type: tea.KeyPgUp}
			m.viewport, cmd = m.viewport.Update(nextmsg)
			return m, cmd
		case key.Matches(msg, _keysPager.PageDown):
			if currentPage >= pageTotal {
				// 下一章
				if m.currentIndex < len(bookDirs)-1 {
					title, content, index := GetBookContent(GetChapterStart(m.currentIndex + 1))
					cmds = append(cmds, pagerCmd(pagerMsg{title: title, content: content, lastPos: GetChapterStart(m.currentIndex + 1), currentIndex: index}))
					return m, tea.Batch(cmds...)
				}
				return m, nil
			}
			currentPage++
			UpdateBookPos(bookName, GetChapterStart(m.currentIndex)+posMapOffset[currentPage])
			nextmsg := tea.KeyMsg{Type: tea.KeyPgDown}
			m.viewport, cmd = m.viewport.Update(nextmsg)
			return m, cmd
		case key.Matches(msg, _keysPager.OpenDir):
			cmds = append(cmds, dirCmd(DirMsg{index: m.currentIndex}))
			cmds = append(cmds, viewCmd(viewDirList))
			return m, tea.Batch(cmds...)
		default:
			return m, nil
		}

	case tea.WindowSizeMsg:

		if !m.ready {
			m.content = proc(m.content, msg.Width, msg.Height-verticalMarginHeight, 0)
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.viewport.MouseWheelEnabled = false
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1

		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	case pagerMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		m.currentIndex = msg.currentIndex
		m.title = msg.title
		m.content = proc(msg.content, winwidth, winheight-verticalMarginHeight, msg.lastPos-GetChapterStart(m.currentIndex))
		m.viewport = viewport.New(winwidth, winheight-verticalMarginHeight)
		m.viewport.YPosition = headerHeight
		m.viewport.SetContent(m.content)
		m.viewport.MouseWheelEnabled = false
		if msg.lastPos > 0 {
			UpdateBookPos(bookName, msg.lastPos)
		} else {
			UpdateBookPos(bookName, GetChapterStart(m.currentIndex)+posMapOffset[currentPage])
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	if jump > 1 {
		for i := 2; i <= jump; i++ {
			msg := tea.KeyMsg{Type: tea.KeyPgDown}
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
		currentPage = jump
		jump = 0
	}

	return m, tea.Batch(cmds...)
}

func (m modelPager) View() string {
	if !m.ready {
		return "\n  Loading..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m modelPager) headerView() string {
	s := "\n"
	s += titleStyle.Render(bookName)
	s += subTitleStyle.Render(m.title)
	s += "\n"
	return s
}

func (m modelPager) footerView() string {
	return "\n" + m.help.View(_keysPager) + "\n"
}

func NewPager() modelPager {
	return modelPager{
		help: help.New(),
	}
}
