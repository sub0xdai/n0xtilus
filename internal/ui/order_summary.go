package ui

import (
    "fmt"
    "strings"
    "github.com/charmbracelet/lipgloss"
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

    // Styles
    titleStyle := lipgloss.NewStyle().
        Foreground(mauve).
        Bold(true).
        Padding(0, 1).
        MarginBottom(1).
        Border(lipgloss.NormalBorder(), false, false, true, false).
        BorderForeground(overlay0)

    sectionStyle := lipgloss.NewStyle().
        Foreground(subtext1).
        Bold(true).
        MarginTop(1).
        MarginBottom(1)

    labelStyle := lipgloss.NewStyle().
        Foreground(subtext1).
        Width(14)

    valueStyle := lipgloss.NewStyle().
        Foreground(text).
        PaddingLeft(1)

    directionStyle := func() lipgloss.Style {
        if o.Direction == "LONG" {
            return lipgloss.NewStyle().
                Foreground(green).
                Bold(true).
                Padding(0, 2)
        }
        return lipgloss.NewStyle().
            Foreground(red).
            Bold(true).
            Padding(0, 2)
    }()

    riskStyle := lipgloss.NewStyle().
        Foreground(peach).
        Bold(true)

    // Build the summary
    var s strings.Builder

    // Title section
    title := lipgloss.JoinHorizontal(
        lipgloss.Center,
        titleStyle.Render("Order Summary"),
        directionStyle.Render(o.Direction),
    )
    s.WriteString(title)
    s.WriteString("\n")

    // Trade details section
    s.WriteString(sectionStyle.Render("Trade Details"))
    s.WriteString("\n")
    tradeDetails := []struct {
        label string
        value string
    }{
        {"Pair", o.Pair},
        {"Entry", fmt.Sprintf("%.2f", o.EntryPrice)},
        {"Stop Loss", fmt.Sprintf("%.2f", o.StopLoss)},
    }

    for _, d := range tradeDetails {
        row := lipgloss.JoinHorizontal(
            lipgloss.Left,
            labelStyle.Render(d.label+":"),
            valueStyle.Render(d.value),
        )
        s.WriteString(row)
        s.WriteString("\n")
    }

    // Position details section
    s.WriteString(sectionStyle.Render("Position Details"))
    s.WriteString("\n")
    positionDetails := []struct {
        label string
        value string
    }{
        {"Leverage", fmt.Sprintf("%.1fx", o.Leverage)},
        {"Position", fmt.Sprintf("%.4f %s", o.Position, strings.Split(o.Pair, "/")[0])},
        {"Risk Amount", riskStyle.Render(fmt.Sprintf("$%.2f", o.RiskAmount))},
    }

    for _, d := range positionDetails {
        row := lipgloss.JoinHorizontal(
            lipgloss.Left,
            labelStyle.Render(d.label+":"),
            valueStyle.Render(d.value),
        )
        s.WriteString(row)
        s.WriteString("\n")
    }

    // Wrap in a box
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(overlay0).
        Padding(1, 2).
        Width(o.width).
        Align(lipgloss.Center)

    return boxStyle.Render(s.String())
}
