package cmd

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Padding(1, 0, 0, 2).Foreground(lipgloss.Color("241"))

type DirectoryPicker struct {
	filepicker   filepicker.Model
	message      string
	selectedFile string
	startingDir  string
	quitting     bool
	err          error
}

func NewDirectoryPicker(dir bool, message string) DirectoryPicker {
	fp := filepicker.New()
	fp.DirAllowed = dir
	fp.FileAllowed = !dir
	h, _ := docStyle.GetFrameSize()
	fp.SetHeight(height - h - 5)
	fp.AutoHeight = true
	fp.ShowHidden = false
	return DirectoryPicker{
		filepicker: fp,
		message:    message,
	}
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (d DirectoryPicker) Init() tea.Cmd {
	return d.filepicker.Init()
}

func (d DirectoryPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SelectedFilePath:
		d.filepicker.CurrentDirectory = string(msg)
		d.startingDir = string(msg)
		return d, d.Init()
	case tea.WindowSizeMsg:
		h, _ := docStyle.GetFrameSize()
		d.filepicker.SetHeight(h)
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			if d.selectedFile != "" {
				fileInfo, err := os.Stat(d.selectedFile)
				if fileInfo.IsDir() != d.filepicker.DirAllowed || err != nil {
					d.err = errors.New(d.selectedFile + " is not valid.")
					d.selectedFile = ""
					return d, clearErrorAfter(2 * time.Second)
				}
				return state.SetAndGetNextCommand(d), tea.Sequence(SendPathCmd(d.startingDir), SendSelectedFile(d.selectedFile))
			}
		case "q", "esc":
			return state.PreviousCommand(), nil
		case "ctrl+c", "ctrl+d":
			d.quitting = true
			return d, tea.Quit
		}
	case clearErrorMsg:
		d.err = nil
	}

	var cmd tea.Cmd
	d.filepicker, cmd = d.filepicker.Update(msg)

	if didSelect, path := d.filepicker.DidSelectFile(msg); didSelect {
		d.selectedFile = path
	}

	if didSelect, path := d.filepicker.DidSelectDisabledFile(msg); didSelect {
		d.err = errors.New(path + " is not valid.")
		d.selectedFile = ""
		return d, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return d, cmd
}

func (d DirectoryPicker) View() string {
	var s strings.Builder
	s.WriteString("\n  ")

	if d.err != nil {
		s.WriteString(d.filepicker.Styles.DisabledFile.Render(d.err.Error()))
	} else {
		s.WriteString(d.message + "\n")
	}
	if d.selectedFile != "" {
		s.WriteString("  Selected: " + d.filepicker.Styles.Selected.Render(d.selectedFile))
	}

	s.WriteString("\n\n" + d.filepicker.View() + "\n")
	s.WriteString(helpStyle.Render("e done • <-/-> move • q quit"))
	return s.String()
}
