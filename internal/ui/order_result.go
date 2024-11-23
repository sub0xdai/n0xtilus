package ui

import (
    "fmt"
    "strings"
    "github.com/charmbracelet/lipgloss"
)

// OrderResult represents the result of a placed order
type OrderResult struct {
    OrderID          string
    Side            string
    Pair            string
    Amount          float64
    Price           float64
    AvailableMargin float64
    width           int
}

// NewOrderResult creates a new order result widget
func NewOrderResult() *OrderResult {
    return &OrderResult{
        width: 60, // default width
    }
}

// SetWidth updates the widget width
func (o *OrderResult) SetWidth(w int) {
    o.width = w
}

// Update updates the order result with new information
func (o *OrderResult) Update(orderID, side, pair string, amount, price, margin float64) {
    o.OrderID = orderID
    o.Side = side
    o.Pair = pair
    o.Amount = amount
    o.Price = price
    o.AvailableMargin = margin
}

// View renders the order result
func (o *OrderResult) View() string {
    if o.OrderID == "" {
        return ""
    }

    // Styles
    successStyle := lipgloss.NewStyle().
        Foreground(green).
        Bold(true)

    labelStyle := lipgloss.NewStyle().
        Foreground(subtext1).
        Width(12)

    valueStyle := lipgloss.NewStyle().
        Foreground(text)

    marginStyle := lipgloss.NewStyle().
        Foreground(peach).
        Bold(true)

    dividerStyle := lipgloss.NewStyle().
        Foreground(overlay0)

    // Build the result view
    var s strings.Builder

    // Success message
    s.WriteString(successStyle.Render("Order Successfully Placed"))
    s.WriteString("\n")
    s.WriteString(dividerStyle.Render(strings.Repeat("─", o.width-4)))
    s.WriteString("\n\n")

    // Order details
    details := []struct {
        label string
        value string
    }{
        {"Order ID", o.OrderID},
        {"Type", fmt.Sprintf("%s %s", strings.ToUpper(o.Side), o.Pair)},
        {"Amount", fmt.Sprintf("%.8f %s", o.Amount, strings.Split(o.Pair, "/")[0])},
        {"Price", fmt.Sprintf("%.2f %s", o.Price, strings.Split(o.Pair, "/")[1])},
        {"Total", fmt.Sprintf("%.2f %s", o.Amount*o.Price, strings.Split(o.Pair, "/")[1])},
    }

    for _, d := range details {
        row := lipgloss.JoinHorizontal(
            lipgloss.Left,
            labelStyle.Render(d.label+":"),
            valueStyle.Render(d.value),
        )
        s.WriteString(row)
        s.WriteString("\n")
    }

    s.WriteString("\n")
    s.WriteString(dividerStyle.Render(strings.Repeat("─", o.width-4)))
    s.WriteString("\n")

    // Margin information
    marginInfo := lipgloss.JoinHorizontal(
        lipgloss.Left,
        labelStyle.Render("Available:"),
        marginStyle.Render(fmt.Sprintf("$%.2f", o.AvailableMargin)),
    )
    s.WriteString(marginInfo)

    // Wrap in a box
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(overlay0).
        Padding(1).
        Width(o.width)

    return boxStyle.Render(s.String())
}
