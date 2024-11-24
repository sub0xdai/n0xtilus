package ui

import (
	"fmt"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sub0xdai/n0xtilus/internal/ui/styles"
)

type ExecuteTradeMsg struct{}

type PositionDashboard struct {
	balance        float64
	positions      []Position
	input          string
	err            string
	width          int
	height         int
	usePlaceholder bool
	helpVisible    bool
}

type Position struct {
	Symbol       string
	Size         float64
	EntryPrice   float64
	CurrentPrice float64
	PnL          float64
}

func NewPositionDashboard(usePlaceholder bool) *PositionDashboard {
	return &PositionDashboard{
		width:          80,
		height:         24,
		usePlaceholder: usePlaceholder,
		helpVisible:    true,
	}
}

func (d *PositionDashboard) Init() tea.Cmd {
	if d.usePlaceholder {
		// Set placeholder data for testing
		d.balance = 10000.0
		d.positions = []Position{
			{
				Symbol:       "BTC/USD",
				Size:        0.5,
				EntryPrice:  50000.0,
				CurrentPrice: 51000.0,
				PnL:         500.0,
			},
		}
	}
	return nil
}

func (d *PositionDashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return d, tea.Quit
		case tea.KeyEnter:
			return d.handleCommand()
		case tea.KeyBackspace, tea.KeyDelete:
			if len(d.input) > 0 {
				d.input = d.input[:len(d.input)-1]
			}
		case tea.KeyEsc:
			d.input = ""
			d.err = ""
		default:
			if msg.Type == tea.KeyRunes {
				d.input += msg.String()
			}
		}
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
	}
	return d, nil
}

func (d *PositionDashboard) handleCommand() (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(strings.ToLower(d.input))
	d.input = ""
	d.err = ""

	switch cmd {
	case "trade", "t":
		d.helpVisible = false
		return d, func() tea.Msg { return ExecuteTradeMsg{} }
	case "help", "h", "?":
		d.helpVisible = !d.helpVisible
	case "clear", "c":
		d.err = ""
	case "quit", "q":
		return d, tea.Quit
	case "":
		// Do nothing for empty command
	default:
		d.err = fmt.Sprintf("Unknown command: %s. Type 'help' for available commands.", cmd)
	}

	return d, nil
}

func (d *PositionDashboard) renderPosition(p Position) string {
	var lines []string

	// Trading pair
	lines = append(lines, styles.PairStyle.Render(p.Symbol))

	// Position details
	positionLine := fmt.Sprintf("%s %s",
		styles.LabelStyle.Render("Position:"),
		styles.ValueStyle.Render(fmt.Sprintf("%.4f BTC ($%.2f)", p.Size, p.EntryPrice)),
	)
	lines = append(lines, positionLine)

	// PnL
	pnlStyle := styles.PnLPositiveStyle
	if p.PnL < 0 {
		pnlStyle = styles.PnLNegativeStyle
	}
	pnlLine := fmt.Sprintf("%s %s",
		styles.LabelStyle.Render("PnL:"),
		pnlStyle.Render(fmt.Sprintf("$%.2f", p.PnL)),
	)
	lines = append(lines, pnlLine)

	// Risk
	riskLine := fmt.Sprintf("%s %s",
		styles.LabelStyle.Render("Risk:"),
		styles.RiskStyle.Render(fmt.Sprintf("$%.2f", p.CurrentPrice)),
	)
	lines = append(lines, riskLine)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (d *PositionDashboard) View() string {
	var sections []string

	// Title box
	titleContent := lipgloss.JoinVertical(lipgloss.Left,
		"",
		fmt.Sprintf("  %s", styles.BalanceStyle.Render(fmt.Sprintf("Balance: $%.2f", d.balance))),
		"",
	)

	titleBox := styles.BoxStyle.Copy().
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(false).
		BorderStyle(lipgloss.NormalBorder()).
		Render(titleContent)

	sections = append(sections, titleBox)

	// Positions box
	var positionsContent string
	if len(d.positions) == 0 {
		positionsContent = "\n  No active positions\n"
	} else {
		var positionSections []string
		positionSections = append(positionSections, "  Active Positions", "")

		for _, p := range d.positions {
			positionSections = append(positionSections, "  "+d.renderPosition(p))
		}
		positionsContent = lipgloss.JoinVertical(lipgloss.Left, positionSections...)
	}

	positionsBox := styles.BoxStyle.Copy().
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		Render(positionsContent)

	sections = append(sections, positionsBox)

	// Help text (if visible)
	if d.helpVisible {
		helpText := []string{
			"",
			"  Commands:",
			"    trade, t    - Open trade input",
			"    help, h, ?  - Toggle help",
			"    clear, c    - Clear messages",
			"    quit, q     - Exit application",
			"    ESC         - Clear input",
			"",
		}
		helpContent := lipgloss.JoinVertical(lipgloss.Left, helpText...)
		helpBox := styles.BoxStyle.Copy().
			BorderStyle(lipgloss.NormalBorder()).
			Render(helpContent)
		sections = append(sections, helpBox)
	}

	// Combine all sections
	dashboard := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add prompt and error at the bottom
	var bottomSections []string

	if d.input != "" {
		prompt := styles.LabelStyle.Render("> " + d.input)
		bottomSections = append(bottomSections, prompt)
	}

	if d.err != "" {
		errorMsg := styles.PnLNegativeStyle.Render(d.err)
		bottomSections = append(bottomSections, errorMsg)
	}

	if len(bottomSections) > 0 {
		bottom := lipgloss.JoinVertical(lipgloss.Left, bottomSections...)
		dashboard = lipgloss.JoinVertical(lipgloss.Left, dashboard, "", bottom)
	}

	return dashboard
}
