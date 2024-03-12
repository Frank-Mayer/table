package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("%v\n", m.table.SelectedRow()),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <filename>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	maxColWidth, err := getMaxColWith()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cols := make([]table.Column, 0, len(records[0]))
	rows := make([]table.Row, 0)
	for _, record := range records {
		rows = append(rows, table.Row(record))
		for i, cell := range record {
			// does the column exist?
			if i >= len(cols) {
				cols = append(cols, table.Column{Title: fmt.Sprintf("%x", i+1)})
			}
			// is the cell length longer than the column?
			if len(cell) > cols[i].Width {
				cols[i].Width = len(cell)
			}
			// is the column longer than the max?
			if cols[i].Width > maxColWidth {
				cols[i].Width = maxColWidth
			}
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func getMaxColWith() (int, error) {
	termWidth := os.Getenv("COLUMNS")
	if termWidth == "" {
		termWidth = "80"
	}
	termWidthInt, err := strconv.Atoi(termWidth)
	if err != nil {
		return 80, err
	}
	return termWidthInt / 2, nil
}
