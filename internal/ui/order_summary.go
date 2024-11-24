package ui

import (
    "fmt"
    "strings"
    "github.com/charmbracelet/lipgloss"
    "github.com/sub0xdai/n0xtilus/internal/ui/styles"
)

// OrderSummary represents an order summary widget
type OrderSummary struct {
    Pair        string
    EntryPrice  float64
    StopLoss    float64
    Leverage    float64
    Direction   string
    RiskAmount  float64
    Position    float64
    width       int
}

// NewOrderSummary creates a new order summary widget
func NewOrderSummary() *OrderSummary {
    return &OrderSummary{
        width: 60, // default width
    }
}

// SetWidth updates the widget width
func (o *OrderSummary) SetWidth(w int) {
    o.width = w
}

// Update updates the order summary with new information
func (o *OrderSummary) Update(pair string, entry, stop, leverage float64, risk, pos float64) {
    o.Pair = pair
    o.EntryPrice = entry
    o.StopLoss = stop
    o.Leverage = leverage
    o.RiskAmount = risk
    o.Position = pos
    o.Direction = func() string {
        if entry > stop {
            return "LONG"
        }
        return "SHORT"
    }()
}

// View renders the order summary
func (o *OrderSummary) View() string {
    if o.Pair == "" {
        return ""
    }

    var content []string

    // Title with direction
    content = append(content, fmt.Sprintf("  Trade Summary (%s)", styles.PairStyle.Render(o.Direction)))
    content = append(content, "")

    // Trading pair
    content = append(content, fmt.Sprintf("  %s %s",
        styles.LabelStyle.Render("Pair:"),
        styles.ValueStyle.Render(o.Pair),
    ))

    // Entry price
    content = append(content, fmt.Sprintf("  %s %s",
        styles.LabelStyle.Render("Entry:"),
        styles.ValueStyle.Render(fmt.Sprintf("$%.2f", o.EntryPrice)),
    ))

    // Stop loss
    content = append(content, fmt.Sprintf("  %s %s",
        styles.LabelStyle.Render("Stop Loss:"),
        styles.ValueStyle.Render(fmt.Sprintf("$%.2f", o.StopLoss)),
    ))

    // Leverage
    content = append(content, fmt.Sprintf("  %s %s",
        styles.LabelStyle.Render("Leverage:"),
        styles.ValueStyle.Render(fmt.Sprintf("%.0fx", o.Leverage)),
    ))

    content = append(content, "")

    // Position size
    content = append(content, fmt.Sprintf("  %s %s",
        styles.LabelStyle.Render("Position:"),
        styles.ValueStyle.Render(fmt.Sprintf("%.4f %s", o.Position, strings.Split(o.Pair, "/")[0])),
    ))

    // Risk amount
    content = append(content, fmt.Sprintf("  %s %s",
        styles.LabelStyle.Render("Risk:"),
        styles.RiskStyle.Render(fmt.Sprintf("$%.2f", o.RiskAmount)),
    ))

    return styles.BoxStyle.Copy().
        BorderStyle(lipgloss.NormalBorder()).
        Render(strings.Join(content, "\n"))
}
