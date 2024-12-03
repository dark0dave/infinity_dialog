package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dark0dave/infinity_dialog/pkg/util"
	"github.com/dark0dave/infinity_file_formats/bg"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type fileview struct {
	title    string
	content  string
	viewport viewport.Model
	callback func() (tea.Model, tea.Cmd)
}

func NewFileView(content string, title string, callback func() (tea.Model, tea.Cmd)) fileview {
	f := fileview{title: title, content: content, callback: callback}
	headerHeight := lipgloss.Height(f.headerView())
	footerHeight := lipgloss.Height(f.footerView())
	verticalMarginHeight := headerHeight + footerHeight
	f.viewport = viewport.New(width, height-verticalMarginHeight)
	f.viewport.YPosition = headerHeight
	f.viewport.SetContent(f.content)
	return f
}

func NewFileViewFromFile(path string, callback func() (tea.Model, tea.Cmd)) fileview {
	content := ""
	if filepath.Ext(path) == ".are" {
		f, err := os.Open(path)
		if err != nil {
		}
		defer f.Close()
		area, err := bg.OpenArea(f)
		if err != nil {
		}
		buf := new(bytes.Buffer)
		err = area.WriteJson(buf)
		if err != nil {
		}
		content = buf.String()
	} else {
		var err error
		content, err = util.ReadFileToString(path)
		if err != nil {
			content = ""
		}
	}
	return NewFileView(content, filepath.Base(path), callback)
}

func (f fileview) Init() tea.Cmd {
	return nil
}

func (f fileview) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return f.callback()
		}

	case tea.WindowSizeMsg:
		setViewport(f, msg)
	}

	f.viewport, cmd = f.viewport.Update(msg)
	return f, cmd
}

func setViewport(f fileview, msg tea.WindowSizeMsg) {
	headerHeight := lipgloss.Height(f.headerView())
	footerHeight := lipgloss.Height(f.footerView())
	verticalMarginHeight := headerHeight + footerHeight
	f.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
	f.viewport.YPosition = headerHeight
	f.viewport.SetContent(f.content)
	f.viewport.Width = msg.Width
	f.viewport.Height = msg.Height - verticalMarginHeight
}

func (f fileview) View() string {
	return fmt.Sprintf("%s\n%s\n%s", f.headerView(), f.viewport.View(), f.footerView())
}

func (f fileview) headerView() string {
	title := titleStyle.Render(f.title)
	line := strings.Repeat("─", max(0, f.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (f fileview) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", f.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, f.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
