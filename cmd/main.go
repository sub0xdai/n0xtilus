package main

import (
	"fmt"
	"log"
	"os"
	"math"
  "github.com/sub0xdai/n0xtilus/internal/config"
	"github.com/sub0xdai/n0xtilus/internal/api"
	"github.com/sub0xdai/n0xtilus/internal/services"
	"github.com/sub0xdai/n0xtilus/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

  // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

	// Initialize API client with placeholder values
	client := api.NewAPIClient(cfg.APIKey, cfg.APISecret, cfg.APIBaseURL)


	// Initialize services
	riskCalculator := services.NewRiskCalculator()
	orderService := services.NewOrderService(client, riskCalculator)

		   

	// Parse the command
  if len(os.Args) < 2 {
        log.Fatal("Usage: n0xtilus <command>")
    }

	command := os.Args[1]
    switch command {
    case "balance":
        handleBalance(client)
    case "trade":
        handleTrade(client, orderService, cfg.RiskPercentage)
    default:
        log.Fatalf("Unknown command: %s", command)
    }
}

func handleBalance(client *api.APIClient) {
    balance, err := client.GetBalance()
    if err != nil {
        log.Fatalf("Failed to get balance: %v", err)
    }
    fmt.Printf("Current balance: %.2f\n", balance)
}

func handleTrade(client *api.APIClient, orderService *services.OrderService, riskPercentage float64) {
    if err := runTradeWidget(client, orderService, riskPercentage); err != nil {
        log.Fatalf("Trade execution failed: %v", err)
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
