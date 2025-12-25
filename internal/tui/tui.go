package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/electr1fy0/sorta/internal"
)

var (
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
)

type model struct {
	dir      string
	ops      []internal.FileOperation
	selected map[int]bool
	cursor   int
	viewport viewport.Model
	quitting bool
	aborted  bool
}

func initialModel(dir string, ops []internal.FileOperation) model {
	selected := make(map[int]bool)
	for i := range ops {
		selected[i] = true
	}

	return model{
		dir:      dir,
		ops:      ops,
		selected: selected,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.aborted = true
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateViewport()
			}
		case "down", "j":
			if m.cursor < len(m.ops)-1 {
				m.cursor++
				m.updateViewport()
			}
		case " ":
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}
			m.updateViewport()
		case "a":
			if len(m.selected) == len(m.ops) {
				m.selected = make(map[int]bool)
			} else {
				for i := range m.ops {
					m.selected[i] = true
				}
			}
			m.updateViewport()
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := 3
		footerHeight := 3
		verticalMarginHeight := headerHeight + footerHeight
		m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
		m.viewport.YPosition = headerHeight
		m.updateViewport()
	}

	return m, nil
}

func (m *model) updateViewport() {
	var sb strings.Builder
	for i, op := range m.ops {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = "[✓]"
		}

		opType := ""
		switch op.OpType {
		case internal.OpMove:
			opType = "MOVE"
		case internal.OpDelete:
			opType = "DEL "
		case internal.OpSkip:
			opType = "SKIP"
		}

		relDest, _ := filepath.Rel(m.dir, op.DestPath)
		if relDest == "" {
			relDest = op.DestPath
		}

		srcName := filepath.Base(op.File.SourcePath)

		line := fmt.Sprintf("%s %s %s %s -> %s", cursor, checked, opType, srcName, relDest)

		if m.cursor == i {
			sb.WriteString(selectedItemStyle.Render(line))
		} else {
			sb.WriteString(itemStyle.Render(line))
		}
		sb.WriteString("\n")
	}
	m.viewport.SetContent(sb.String())

	if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	} else if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	}
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	header := titleStyle.Render("Review Operations") + "\n"
	help := helpStyle.Render("↑/↓: move • space: toggle • a: toggle all • enter: confirm • q: cancel")

	if m.viewport.Width == 0 {
		return "Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), help)
}

func SelectOperations(dir string, ops []internal.FileOperation) ([]internal.FileOperation, error) {
	p := tea.NewProgram(initialModel(dir, ops))
	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	finalModel := m.(model)
	if finalModel.aborted {
		return nil, fmt.Errorf("operation cancelled by user")
	}

	var selected []internal.FileOperation
	for i, op := range ops {
		if finalModel.selected[i] {
			selected = append(selected, op)
		}
	}

	return selected, nil
}
