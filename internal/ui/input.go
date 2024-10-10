package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TradeStep int

const (
	StepSelectPair TradeStep = iota
	StepEntryPrice
	StepStopLoss
	StepLeverage
	StepConfirmation
	StepComplete
)

type TradeInputWidget struct {
	pairs           []string
	currentStep     TradeStep
	inputs          []textinput.Model
	err             error
	tradeInfo       map[string]string
	confirmationMsg string
	Confirmed       bool // Changed to public
}

func NewTradeInputWidget(pairs []string) *TradeInputWidget {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 32
		inputs[i] = t
	}
	inputs[0].Placeholder = "Select pair (1-5)"
	inputs[1].Placeholder = "Entry price"
	inputs[2].Placeholder = "Stop loss"
	inputs[3].Placeholder = "Leverage"
	inputs[0].Focus()
	return &TradeInputWidget{
		pairs:       pairs,
		currentStep: StepSelectPair,
		inputs:      inputs,
	}
}

func (m *TradeInputWidget) Init() tea.Cmd {
	return textinput.Blink
}

func (m *TradeInputWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyCtrlC:
            return m, tea.Quit
        case tea.KeyEnter:
            if m.currentStep == StepConfirmation {
                m.Confirmed = strings.ToLower(strings.TrimSpace(m.confirmationMsg)) == "y"
                m.currentStep = StepComplete
                return m, tea.Quit
            }
            return m, m.nextStep()
        }
      case cursor.BlinkMsg:
        return m, nil
      default:
        return m, nil
    }

    var cmd tea.Cmd
    if m.currentStep == StepConfirmation {
        m.confirmationMsg += msg.(tea.KeyMsg).String()
    } else {
        m.inputs[m.currentStep], cmd = m.inputs[m.currentStep].Update(msg)
    }
    return m, cmd
}

func (m *TradeInputWidget) nextStep() tea.Cmd {
    switch m.currentStep {
    case StepSelectPair, StepEntryPrice, StepStopLoss:
        m.currentStep++
        return m.inputs[m.currentStep].Focus()
    case StepLeverage:
        m.currentStep = StepConfirmation
        m.calculateTradeInfo()
    }
    return nil
}

func (m *TradeInputWidget) calculateTradeInfo() {
	pairIndex, _ := strconv.Atoi(m.inputs[0].Value())
	entryPrice, _ := strconv.ParseFloat(m.inputs[1].Value(), 64)
	stopLoss, _ := strconv.ParseFloat(m.inputs[2].Value(), 64)
	leverage, _ := strconv.ParseFloat(m.inputs[3].Value(), 64)

	m.tradeInfo = map[string]string{
		"Pair":             m.pairs[pairIndex-1],
		"Entry Price":      fmt.Sprintf("$%.2f", entryPrice),
		"Stop Loss":        fmt.Sprintf("$%.2f", stopLoss),
		"Leverage":         fmt.Sprintf("%.2fx", leverage),
		"Position Size":    "$5000.00", // Placeholder
		"Margin Required":  "$500.00",  // Placeholder
		"Available Margin": "$1000.00", // Placeholder
	}
}

func (m *TradeInputWidget) View() string {
	var s string
	switch m.currentStep {
	case StepSelectPair:
		s += "Select trading pair:\n"
		for i, pair := range m.pairs {
			s += fmt.Sprintf("%d. %s\n", i+1, pair)
		}
		s += "\n" + m.inputs[0].View()
	case StepEntryPrice:
		s += "Enter the entry price:\n\n" + m.inputs[1].View()
	case StepStopLoss:
		s += "Enter the stop loss price:\n\n" + m.inputs[2].View()
	case StepLeverage:
		s += "Enter the leverage (e.g., 5 for 5x leverage):\n\n" + m.inputs[3].View()
	case StepConfirmation:
		s += "Trade Information\n\n"
		for k, v := range m.tradeInfo {
			s += fmt.Sprintf("%s: %s\n", k, v)
		}
		s += "\nDo you want to place this trade? (y/n): " + m.confirmationMsg
	case StepComplete:
		if m.Confirmed {
			s += "Trade placed successfully!"
		} else {
			s += "Trade cancelled."
		}
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Render(s)
}

func (m *TradeInputWidget) IsComplete() bool {
	return m.currentStep == StepComplete
}

func (m *TradeInputWidget) GetInputs() (int, float64, float64, float64, error) {
	pairIndex, err := strconv.Atoi(m.inputs[0].Value())
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid pair selection: %v", err)
	}
	pairIndex-- // Adjust for 0-based index

	entry, err := strconv.ParseFloat(m.inputs[1].Value(), 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid entry price: %v", err)
	}

	stopLoss, err := strconv.ParseFloat(m.inputs[2].Value(), 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid stop loss: %v", err)
	}

	leverage, err := strconv.ParseFloat(m.inputs[3].Value(), 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid leverage: %v", err)
	}

	return pairIndex, entry, stopLoss, leverage, nil
}
