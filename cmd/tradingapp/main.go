package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

	positionSize, err := orderService.CalculatePositionSize(riskPercentage, entryPrice, stopLossPrice)
	if err != nil {
		log.Fatalf("Failed to calculate position size: %v", err)
	}

	fmt.Printf("Calculated position size: %f\n", positionSize)

	fmt.Print("Do you want to place this trade? (yes/no): ")
	confirmStr, _ := reader.ReadString('\n')
	confirm := strings.TrimSpace(confirmStr)

	if strings.ToLower(confirm) == "yes" {
		// Place the order
		order, err := orderService.PlaceOrder(selectedPair, "buy", fmt.Sprintf("%f", positionSize), fmt.Sprintf("%f", entryPrice))
		if err != nil {
			log.Fatalf("Failed to place order: %v", err)
		}
		fmt.Printf("Order placed: %s\n", order)
	} else {
		fmt.Println("Trade cancelled.")
	}
}
