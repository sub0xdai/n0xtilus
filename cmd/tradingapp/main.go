package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"math"
	"strconv"
	"strings"
	"github.com/tehuticode/b0xfi/internal/api"
	"github.com/tehuticode/b0xfi/internal/services"
)

func main() {
	// Initialize API client with placeholder values
	client := api.NewAPIClient("placeholder_key", "placeholder_secret")
	// Initialize services
	riskCalculator := services.NewRiskCalculator()
	orderService := services.NewOrderService(client, riskCalculator)
	// Check if a command was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: b0xfi <command>")
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
		executeTrade(client, orderService, 1.0) // Using a placeholder risk percentage of 1%
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: balance, trade")
		os.Exit(1)
	}
}

func executeTrade(client *api.APIClient, orderService *services.OrderService, riskPercentage float64) {
    reader := bufio.NewReader(os.Stdin)
    
    // Get account balance first (this is our available margin)
    availableMargin, err := client.GetBalance()
    if err != nil {
        log.Fatalf("Failed to get balance: %v", err)
    }
    fmt.Printf("Available margin: $%.2f\n", availableMargin)

    // Fetch tradable pairs
    pairs, err := client.GetTradablePairs()
    if err != nil {
        log.Fatalf("Failed to fetch tradable pairs: %v", err)
    }
    
    // Display pairs and let user select
    fmt.Println("Available trading pairs:")
    for i, pair := range pairs {
        fmt.Printf("%d. %s\n", i+1, pair)
    }
    var selectedPair string
    for {
        fmt.Print("Enter the number of the pair you want to trade: ")
        input, _ := reader.ReadString('\n')
        index, err := strconv.Atoi(strings.TrimSpace(input))
        if err != nil || index < 1 || index > len(pairs) {
            fmt.Println("Invalid selection. Please try again.")
            continue
        }
        selectedPair = pairs[index-1]
        break
    }

    fmt.Print("Enter the entry price: ")
    entryPriceStr, _ := reader.ReadString('\n')
    entryPrice, err := strconv.ParseFloat(strings.TrimSpace(entryPriceStr), 64)
    if err != nil {
        log.Fatalf("Invalid entry price: %v", err)
    }

    fmt.Print("Enter the stop loss price: ")
    stopLossPriceStr, _ := reader.ReadString('\n')
    stopLossPrice, err := strconv.ParseFloat(strings.TrimSpace(stopLossPriceStr), 64)
    if err != nil {
        log.Fatalf("Invalid stop loss price: %v", err)
    }

    fmt.Print("Enter the leverage (e.g., 5 for 5x leverage): ")
    leverageStr, _ := reader.ReadString('\n')
    leverage, err := strconv.ParseFloat(strings.TrimSpace(leverageStr), 64)
    if err != nil {
        log.Fatalf("Invalid leverage: %v", err)
    }
    
    // Calculate the maximum risk amount based on the risk percentage
    maxRiskAmount := availableMargin * riskPercentage
    
    // Calculate the position size based on the risk amount and stop loss
    stopLossDistance := math.Abs(entryPrice - stopLossPrice) / entryPrice // as a percentage
    maxPositionSize := maxRiskAmount / stopLossDistance

    // Apply leverage to get the maximum leveraged position size
    maxLeveragedPositionSize := maxPositionSize * leverage

    // Calculate the margin required for this position
    marginRequired := maxLeveragedPositionSize / leverage

    // Check if we have enough margin available
    if marginRequired > availableMargin {
        fmt.Printf("Warning: Required margin ($%.2f) exceeds available margin ($%.2f)\n", marginRequired, availableMargin)
        fmt.Println("Adjusting position size to match available margin...")
        marginRequired = availableMargin
        maxLeveragedPositionSize = marginRequired * leverage
    }

    fmt.Printf("Calculated position size: $%.2f\n", maxLeveragedPositionSize)
    fmt.Printf("Margin used: $%.2f\n", marginRequired)
    fmt.Printf("Effective leverage: %.2fx\n", maxLeveragedPositionSize / marginRequired)

    fmt.Print("Do you want to place this trade? (yes/no): ")
    confirmStr, _ := reader.ReadString('\n')
    confirm := strings.TrimSpace(confirmStr)
    if strings.ToLower(confirm) == "yes" {
        // Place the order
        order, err := orderService.PlaceOrder(selectedPair, "buy", fmt.Sprintf("%.2f", maxLeveragedPositionSize), fmt.Sprintf("%.2f", entryPrice))
        if err != nil {
            log.Fatalf("Failed to place order: %v", err)
        }
        fmt.Printf("Order placed: %s\n", order)
        
        // Update available margin
        availableMargin -= marginRequired
        fmt.Printf("Remaining available margin: $%.2f\n", availableMargin)
    } else {
        fmt.Println("Trade cancelled.")
    }
}
