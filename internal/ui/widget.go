package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ... (keep existing Main struct and methods)

type TradeInfoWidget struct {
	info   map[string]string
	style  lipgloss.Style
	done   bool
}

func NewTradeInfoWidget(info map[string]string) *TradeInfoWidget {
	return &TradeInfoWidget{
		info:  info,
		style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1),
	}
}

func (m *TradeInfoWidget) Init() tea.Cmd {
	return nil
}

func (m *TradeInfoWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *TradeInfoWidget) View() string {
	var b strings.Builder
	b.WriteString("Trade Information\n\n")
	
	for k, v := range m.info {
		b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	
	b.WriteString("\nPress Enter to continue")
	
	return m.style.Render(b.String())
}

type ConfirmWidget struct {
	question string
	confirmed bool
	done     bool
}

func NewConfirmWidget(question string) *ConfirmWidget {
	return &ConfirmWidget{
		question: question,
	}
}

func (m *ConfirmWidget) Init() tea.Cmd {
	return nil
}

func (m *ConfirmWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.confirmed = true
			m.done = true
			return m, tea.Quit
		case "n", "N":
			m.confirmed = false
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ConfirmWidget) View() string {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Render(
		fmt.Sprintf("%s (y/n)", m.question),
	)
}

func (m *ConfirmWidget) Confirmed() bool {
	return m.confirmed
}
