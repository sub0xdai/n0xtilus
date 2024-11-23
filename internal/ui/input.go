package ui

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha colors
var (
    rosewater = lipgloss.Color("#f5e0dc")
    flamingo  = lipgloss.Color("#f2cdcd")
    pink      = lipgloss.Color("#f5c2e7")
    mauve     = lipgloss.Color("#cba6f7")
    red       = lipgloss.Color("#f38ba8")
    maroon    = lipgloss.Color("#eba0ac")
    peach     = lipgloss.Color("#fab387")
    yellow    = lipgloss.Color("#f9e2af")
    green     = lipgloss.Color("#a6e3a1")
    teal      = lipgloss.Color("#94e2d5")
    sky       = lipgloss.Color("#89dceb")
    sapphire  = lipgloss.Color("#74c7ec")
    blue      = lipgloss.Color("#89b4fa")
    lavender  = lipgloss.Color("#b4befe")
    text      = lipgloss.Color("#cdd6f4")
    subtext1  = lipgloss.Color("#bac2de")
    overlay0  = lipgloss.Color("#6c7086")
    surface0  = lipgloss.Color("#313244")
    base      = lipgloss.Color("#1e1e2e")
    mantle    = lipgloss.Color("#181825")
)

// Styles
var (
    titleStyle = lipgloss.NewStyle().
        Foreground(mauve).
        Bold(true).
        Padding(1, 0, 0, 0)

    promptStyle = lipgloss.NewStyle().
        Foreground(lavender)

    infoStyle = lipgloss.NewStyle().
        Foreground(subtext1)

    errorStyle = lipgloss.NewStyle().
        Foreground(red)

    inputStyle = lipgloss.NewStyle().
        Foreground(text).
        Background(surface0).
        Padding(0, 1)

    summaryStyle = lipgloss.NewStyle().
        Foreground(green).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(overlay0).
        Padding(1)

    pairStyle = lipgloss.NewStyle().
        Foreground(peach)

    confirmStyle = lipgloss.NewStyle().
        Foreground(sky).
        Bold(true)

    containerStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(overlay0).
        Padding(1).
        Align(lipgloss.Center)
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
        t.Cursor.Style = lipgloss.NewStyle().Foreground(sky)
        t.PromptStyle = lipgloss.NewStyle().Foreground(lavender)
        t.TextStyle = lipgloss.NewStyle().Foreground(text)
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
    var s strings.Builder

    // Title
    s.WriteString(titleStyle.Render("n0xtilus Trade Entry"))
    s.WriteString("\n\n")

    if m.err != nil {
        s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
        s.WriteString("\n\n")
    }

    switch m.currentStep {
    case StepPair:
        s.WriteString(promptStyle.Render("Select trading pair:"))
        s.WriteString("\n")
        for i, pair := range m.pairs {
            s.WriteString(pairStyle.Render(fmt.Sprintf("%d. %s\n", i+1, pair)))
        }
        s.WriteString("\n")
        s.WriteString(inputStyle.Render(m.inputs[0].View()))

    case StepEntryPrice:
        s.WriteString(promptStyle.Render("Enter entry price:"))
        s.WriteString("\n")
        s.WriteString(inputStyle.Render(m.inputs[1].View()))

    case StepStopLoss:
        s.WriteString(promptStyle.Render("Enter stop loss price:"))
        s.WriteString("\n")
        s.WriteString(inputStyle.Render(m.inputs[2].View()))

    case StepLeverage:
        s.WriteString(promptStyle.Render("Enter leverage (1-100x):"))
        s.WriteString("\n")
        s.WriteString(inputStyle.Render(m.inputs[3].View()))

    case StepConfirmation:
        // Use the new order summary widget
        s.WriteString(m.summary.View())
        s.WriteString("\n")
        s.WriteString(confirmStyle.Render("Confirm trade? (y/n): "))
    }

    // Center the content using actual window width
    return containerStyle.Width(min(m.width-4, 60)).Render(s.String())
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
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
