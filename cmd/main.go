package main

import (
	"fmt"
	"log"
	"os"
	"math"
	"github.com/sub0xdai/n0xtilus/internal/config"
	"github.com/sub0xdai/n0xtilus/internal/api"
	"github.com/sub0xdai/n0xtilus/internal/services"
	"github.com/sub0xdai/n0xtilus/internal/services/risk_calculator"
	"github.com/sub0xdai/n0xtilus/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Config loaded - TestMode: %v", cfg.TestMode)

	if !cfg.TestMode {
		if cfg.APIKey == "" || cfg.APISecret == "" || cfg.APIBaseURL == "" {
			log.Fatal("API credentials not configured. Please update config.yaml")
		}
	}

	if cfg.RiskPercentage <= 0 || cfg.RiskPercentage > 100 {
		log.Fatal("Risk percentage must be between 0 and 100")
	}

	// Initialize API client
	client := api.NewAPIClient(cfg.APIKey, cfg.APISecret, cfg.APIBaseURL)
	if client == nil {
		log.Fatal("Failed to initialize API client")
	}

	// Initialize services
	riskCalc := risk_calculator.NewRiskCalculator()
	if riskCalc == nil {
		log.Fatal("Failed to initialize risk calculator")
	}

	orderService := services.NewOrderService(client, riskCalc)
	if orderService == nil {
		log.Fatal("Failed to initialize order service")
	}

	// Parse command
	if len(os.Args) < 2 {
		log.Fatal("Usage: n0xtilus <command>\n\nAvailable commands:\n  balance - Show account balance\n  trade   - Open trading interface")
	}

	command := os.Args[1]
	switch command {
	case "balance":
		handleBalance(client)
	case "trade":
		handleTrade(client, orderService, cfg.RiskPercentage/100) // Convert percentage to decimal
	default:
		log.Fatalf("Unknown command: %s\n\nAvailable commands:\n  balance - Show account balance\n  trade   - Open trading interface", command)
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
    if client == nil {
        return fmt.Errorf("API client is not initialized")
    }

    if orderService == nil {
        return fmt.Errorf("order service is not initialized")
    }

    availableMargin, err := client.GetBalance()
    if err != nil {
        return fmt.Errorf("failed to get balance: %v", err)
    }

    pairs, err := client.GetTradablePairs()
    if err != nil {
        return fmt.Errorf("failed to fetch tradable pairs: %v", err)
    }

    if len(pairs) == 0 {
        return fmt.Errorf("no tradable pairs available")
    }

    tradeInput := ui.NewTradeInputWidget(pairs)
    if tradeInput == nil {
        return fmt.Errorf("failed to create trade input widget")
    }

    p := tea.NewProgram(tradeInput)
    if p == nil {
        return fmt.Errorf("failed to create bubbletea program")
    }

    finalModel, err := p.Run()
    if err != nil {
        return fmt.Errorf("failed to run trade input: %v", err)
    }

    if finalModel == nil {
        return fmt.Errorf("trade input returned nil model")
    }

    finalTradeInput, ok := finalModel.(*ui.TradeInputWidget)
    if !ok {
        return fmt.Errorf("invalid model type returned")
    }

    if !finalTradeInput.IsComplete() {
        return fmt.Errorf("trade input process was not completed")
    }

    pairIndex, entryPrice, stopLossPrice, leverage, err := finalTradeInput.GetInputs()
    if err != nil {
        return fmt.Errorf("failed to get trade inputs: %v", err)
    }

    if pairIndex < 0 || pairIndex >= len(pairs) {
        return fmt.Errorf("invalid pair index: %d", pairIndex)
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

        // Create and display order result
        orderResult := ui.NewOrderResult()
        orderResult.SetWidth(80) // Set a comfortable width
        orderResult.Update(
            order, // Using the order ID/string from the API
            "buy",
            selectedPair,
            leveragedPositionSize,
            entryPrice,
            availableMargin - marginRequired,
        )

        // Clear screen and show result
        fmt.Print("\033[H\033[2J") // Clear screen
        fmt.Println(orderResult.View())
    } else {
        fmt.Println("Trade cancelled.")
    }

    return nil
}
