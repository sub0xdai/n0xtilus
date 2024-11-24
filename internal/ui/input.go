package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sub0xdai/n0xtilus/internal/ui/styles"
)

type InputStep int

const (
	StepPair InputStep = iota
	StepEntryPrice
	StepStopLoss
	StepLeverage
	StepConfirmation
	StepComplete
)

type TradeInputWidget struct {
	pairs       []string
	currentStep InputStep
	inputs      []textinput.Model
	err         error
	tradeInfo   map[string]string
	Confirmed   bool
	width       int
	height      int
	summary     *OrderSummary
}

func NewTradeInputWidget(pairs []string) *TradeInputWidget {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(styles.Sky)
		t.PromptStyle = lipgloss.NewStyle().Foreground(styles.Lavender)
		t.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
		t.CharLimit = 32
		inputs[i] = t
	}

	inputs[0].Placeholder = "Enter number (1-5)"
	inputs[1].Placeholder = "0.00"
	inputs[2].Placeholder = "0.00"
	inputs[3].Placeholder = "1-100"

	inputs[0].Focus()

	return &TradeInputWidget{
		pairs:       pairs,
		currentStep: StepPair,
		inputs:      inputs,
		tradeInfo:   make(map[string]string),
		width:       80,  // Default width
		height:      24,  // Default height
		summary:     NewOrderSummary(),
	}
}

func (m *TradeInputWidget) Init() tea.Cmd {
	return textinput.Blink
}

func (m *TradeInputWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.summary.SetWidth(min(m.width-4, 60))
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.currentStep == StepConfirmation {
				switch msg.String() {
				case "y", "Y":
					m.Confirmed = true
					m.currentStep = StepComplete
					return m, tea.Quit
				case "n", "N":
					m.currentStep = StepComplete
					return m, tea.Quit
				default:
					return m, nil
				}
			}
			return m, m.nextStep()
		}
	}

	// Handle input for confirmation step
	if m.currentStep == StepConfirmation {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				m.Confirmed = true
				m.currentStep = StepComplete
				return m, tea.Quit
			case "n", "N":
				m.currentStep = StepComplete
				return m, tea.Quit
			}
		}
		return m, nil
	}

	// Only update input if we're in an input step
	if m.currentStep < StepConfirmation {
		var cmd tea.Cmd
		m.inputs[m.currentStep], cmd = m.inputs[m.currentStep].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *TradeInputWidget) nextStep() tea.Cmd {
	switch m.currentStep {
	case StepPair, StepEntryPrice, StepStopLoss:
		m.currentStep++
		return m.inputs[m.currentStep].Focus()
	case StepLeverage:
		if err := m.validateInputs(); err != nil {
			m.err = err
			return nil
		}
		m.currentStep = StepConfirmation
		m.calculateTradeInfo()
	}
	return nil
}

func (m *TradeInputWidget) validateInputs() error {
	// Validate pair selection
	pairNum, err := strconv.Atoi(m.inputs[0].Value())
	if err != nil || pairNum < 1 || pairNum > len(m.pairs) {
		return fmt.Errorf("invalid pair selection: must be between 1 and %d", len(m.pairs))
	}

	// Validate entry price
	if _, err := strconv.ParseFloat(m.inputs[1].Value(), 64); err != nil {
		return fmt.Errorf("invalid entry price: must be a number")
	}

	// Validate stop loss
	if _, err := strconv.ParseFloat(m.inputs[2].Value(), 64); err != nil {
		return fmt.Errorf("invalid stop loss: must be a number")
	}

	// Validate leverage
	leverage, err := strconv.ParseFloat(m.inputs[3].Value(), 64)
	if err != nil || leverage <= 0 || leverage > 100 {
		return fmt.Errorf("invalid leverage: must be between 1 and 100")
	}

	return nil
}

func (m *TradeInputWidget) calculateTradeInfo() {
	pairIdx, _ := strconv.Atoi(m.inputs[0].Value())
	entryPrice, _ := strconv.ParseFloat(m.inputs[1].Value(), 64)
	stopLoss, _ := strconv.ParseFloat(m.inputs[2].Value(), 64)
	leverage, _ := strconv.ParseFloat(m.inputs[3].Value(), 64)

	// Calculate risk and position size (example values)
	riskAmount := 100.0 // This should come from your risk calculator
	position := 0.5     // This should come from your position size calculator

	// Update order summary
	m.summary.Update(
		m.pairs[pairIdx-1],
		entryPrice,
		stopLoss,
		leverage,
		riskAmount,
		position,
	)

	// Keep the old trade info for backward compatibility
	m.tradeInfo = map[string]string{
		"Pair":        m.pairs[pairIdx-1],
		"Entry Price": fmt.Sprintf("%.2f", entryPrice),
		"Stop Loss":   fmt.Sprintf("%.2f", stopLoss),
		"Leverage":    fmt.Sprintf("%.1fx", leverage),
		"Direction":   func() string {
			if entryPrice > stopLoss {
				return "LONG"
			}
			return "SHORT"
		}(),
	}
}

func (m *TradeInputWidget) View() string {
	var sections []string

	// Title box
	titleBox := styles.BoxStyle.Copy().
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(false).
		Render(fmt.Sprintf("\n  %s\n", styles.TitleStyle.Render("n0xtilus Trade Entry")))
	sections = append(sections, titleBox)

	// Content box
	var content []string
	if m.err != nil {
		content = append(content, styles.PnLNegativeStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		content = append(content, "")
	}

	switch m.currentStep {
	case StepPair:
		content = append(content, "  Select trading pair:")
		content = append(content, "")
		for i, pair := range m.pairs {
			content = append(content, fmt.Sprintf("    %d. %s", i+1, styles.PairStyle.Render(pair)))
		}
		content = append(content, "")
		content = append(content, fmt.Sprintf("  > %s", m.inputs[0].View()))

	case StepEntryPrice:
		content = append(content, "  Enter entry price:")
		content = append(content, "")
		content = append(content, fmt.Sprintf("  > %s", m.inputs[1].View()))

	case StepStopLoss:
		content = append(content, "  Enter stop loss price:")
		content = append(content, "")
		content = append(content, fmt.Sprintf("  > %s", m.inputs[2].View()))

	case StepLeverage:
		content = append(content, "  Enter leverage (1-100x):")
		content = append(content, "")
		content = append(content, fmt.Sprintf("  > %s", m.inputs[3].View()))

	case StepConfirmation:
		// Use the new order summary widget
		content = append(content, m.summary.View())
		content = append(content, "")
		content = append(content, fmt.Sprintf("  %s", styles.ConfirmStyle.Render("Confirm trade? (y/n): ")))
	}

	contentBox := styles.BoxStyle.Copy().
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		Render(strings.Join(content, "\n"))
	
	sections = append(sections, contentBox)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *TradeInputWidget) IsComplete() bool {
	return m.currentStep == StepComplete
}

func (m *TradeInputWidget) GetInputs() (int, float64, float64, float64, error) {
	pairNum, err := strconv.Atoi(m.inputs[0].Value())
	if err != nil || pairNum < 1 || pairNum > len(m.pairs) {
		return 0, 0, 0, 0, fmt.Errorf("invalid pair selection: must be between 1 and %d", len(m.pairs))
	}

	entry, err := strconv.ParseFloat(m.inputs[1].Value(), 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid entry price: %v", err)
	}

	stopLoss, err := strconv.ParseFloat(m.inputs[2].Value(), 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid stop loss: %v", err)
	}

	leverage, err := strconv.ParseFloat(m.inputs[3].Value(), 64)
	if err != nil || leverage <= 0 || leverage > 100 {
		return 0, 0, 0, 0, fmt.Errorf("invalid leverage: must be between 1 and 100")
	}

	return pairNum - 1, entry, stopLoss, leverage, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
