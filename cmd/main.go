package main

import (
	"fmt"
	"log"
	"os"
	"math"
	"github.com/sub0xdai/n0xtilus/internal/api"
	"github.com/sub0xdai/n0xtilus/internal/services"
	"github.com/sub0xdai/n0xtilus/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize API client with placeholder values
	client := api.NewAPIClient("placeholder_key", "placeholder_secret")
	// Initialize services
	riskCalculator := services.NewRiskCalculator()
	orderService := services.NewOrderService(client, riskCalculator)

	// Check if a command was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: n0xtilus <command>")
		fmt.Println("Available commands: balance, trade")
		os.Exit(1)
	}

	// Parse the command
	command := os.Args[1]
	switch command {
	case "balance":
		balance, err := client.GetBalance()
		if err != nil {
			log.Fatalf("Failed to get balance: %v", err)
		}
		fmt.Printf("Current balance: %f\n", balance)
	case "trade":
		if err := runTradeWidget(client, orderService, 0.02); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: balance, trade")
		os.Exit(1)
	}
}

func runTradeWidget(client *api.APIClient, orderService *services.OrderService, riskPercentage float64) error {
	availableMargin, err := client.GetBalance()
	if err != nil {
		return fmt.Errorf("failed to get balance: %v", err)
	}

	pairs, err := client.GetTradablePairs()
	if err != nil {
		return fmt.Errorf("failed to fetch tradable pairs: %v", err)
	}

	tradeInput := ui.NewTradeInputWidget(pairs)
	p := tea.NewProgram(tradeInput)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run trade input: %v", err)
	}

	finalTradeInput := finalModel.(*ui.TradeInputWidget)
	if !finalTradeInput.IsComplete() {
		return fmt.Errorf("trade input process was not completed")
	}

	pairIndex, entryPrice, stopLossPrice, leverage, err := finalTradeInput.GetInputs()
	if err != nil {
		return err
	}

	selectedPair := pairs[pairIndex]

	// Calculate trade parameters
	maxRiskAmount := availableMargin * riskPercentage
	stopLossDistance := math.Abs(entryPrice - stopLossPrice) / entryPrice
	positionSize := maxRiskAmount / stopLossDistance
	leveragedPositionSize := positionSize * leverage
	marginRequired := leveragedPositionSize / leverage

	if marginRequired > availableMargin {
		return fmt.Errorf("required margin ($%.2f) exceeds available margin ($%.2f)", marginRequired, availableMargin)
	}

	if finalTradeInput.Confirmed {
		// Place the order
		order, err := orderService.PlaceOrder(selectedPair, "buy", fmt.Sprintf("%.2f", leveragedPositionSize), fmt.Sprintf("%.2f", entryPrice))
		if err != nil {
			return fmt.Errorf("failed to place order: %v", err)
		}
		fmt.Printf("Order placed: %s\n", order)
		
		// Update available margin
		availableMargin -= marginRequired
		fmt.Printf("Remaining available margin: $%.2f\n", availableMargin)
	} else {
		fmt.Println("Trade cancelled.")
	}

	return nil
}
