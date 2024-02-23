package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type table struct {
	list     list.Model
	delegate func(grepResult)
}

func (t table) Init() tea.Cmd {
	return nil
}

func (t table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return t, tea.Quit
		} else if msg.String() == "enter" {
			if t.delegate != nil {
				gr, ok := t.list.SelectedItem().(grepResult)
				if ok {
					t.delegate(gr)
				}
			}
			return t, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return t, cmd
}

func (t table) View() string {
	return docStyle.Render(t.list.View())
}

func main() {
	results, err := execute(strings.Join(os.Args[1:], " "))
	if err != nil {
		log.Fatal(err.Error())
	}

	l := list.New(results, list.NewDefaultDelegate(), 0, 0)
	l.Title = strings.Join(os.Args[1:], " ")
	m := table{
		list: l,
		delegate: func(gr grepResult) {
			tea.ClearScreen()
			err := openVim(gr)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err.Error())
	}
}

func execute(s string) ([]list.Item, error) {
	cmd := exec.Command("grep", "-r", "-n", s, ".")
	out, err := cmd.Output()
	if string(out) == "" {
		if err != nil && err.Error() != "exit status 1" {
			return nil, err
		}
	}

	results := make([]list.Item, 0)
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}

		p := strings.Split(line, ":")
		if len(p) < 3 {
			continue
		}

		result := grepResult{
			FileName:   p[0],
			LineNumber: p[1],
			Contents:   strings.Join(p[2:], " "),
		}

		results = append(results, result)
	}

	return results, nil
}
func openVim(gr grepResult) error {
	cmd := exec.Command("vim", "+"+gr.LineNumber, gr.FileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

type grepResult struct {
	FileName   string
	LineNumber string
	Contents   string
}

func (gr grepResult) Title() string {
	return gr.Contents
}

func (gr grepResult) Description() string {
	return gr.FileName + ":" + gr.LineNumber
}

func (gr grepResult) FilterValue() string {
	return gr.Contents
}
