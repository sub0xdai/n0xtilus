package main

import (
	"fmt"
	"log"
	"os"
	"math"
	"strconv"
	"strings"
	"time"
	"github.com/tehuticode/n0xtilus/internal/api"
	"github.com/tehuticode/n0xtilus/internal/services"
	"github.com/tehuticode/n0xtilus/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	errorLog *log.Logger
)

func initErrorLog() {
	currentTime := time.Now()
	logFileName := fmt.Sprintf("error_log_%s.txt", currentTime.Format("2006-01-02"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	errorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func logError(err error) {
	if errorLog != nil {
		errorLog.Println(err)
	}
}

func main() {
	initErrorLog()

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
			logError(fmt.Errorf("failed to get balance: %v", err))
			log.Fatalf("Failed to get balance: %v", err)
		}
		fmt.Printf("Current balance: %f\n", balance)
	case "trade":
		if err := runTradeWidget(client, orderService, 0.02); err != nil {
			logError(err)
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

	pairsQuestion := fmt.Sprintf("Select trading pair (1-%d):\n", len(pairs))
	for i, pair := range pairs {
		pairsQuestion += fmt.Sprintf("%d. %s\n", i+1, pair)
	}

	questions := []ui.Question{
		ui.NewShortQuestion(pairsQuestion),
		ui.NewShortQuestion("Entry:"),
		ui.NewShortQuestion("Stop Loss:"),
		ui.NewShortQuestion("Leverage (5 for 5x leverage):"),
	}

	widget := ui.New(questions)
	p := tea.NewProgram(widget, tea.WithAltScreen())

	m, err := p.Run()
	if err != nil {
		return err
	}

	answers := m.(*ui.Main).GetAnswers()

	// Process answers
	pairIndex, err := strconv.Atoi(answers[0])
	if err != nil {
		return fmt.Errorf("invalid pair selection: %v", err)
	}
	selectedPair := pairs[pairIndex-1]
	entryPrice, err := strconv.ParseFloat(answers[1], 64)
	if err != nil {
		return fmt.Errorf("invalid entry price: %v", err)
	}
	stopLossPrice, err := strconv.ParseFloat(answers[2], 64)
	if err != nil {
		return fmt.Errorf("invalid stop loss price: %v", err)
	}
	leverage, err := strconv.ParseFloat(answers[3], 64)
	if err != nil {
		return fmt.Errorf("invalid leverage: %v", err)
	}

	// Calculate trade parameters
	maxRiskAmount := availableMargin * riskPercentage
	stopLossDistance := math.Abs(entryPrice - stopLossPrice) / entryPrice
	positionSize := maxRiskAmount / stopLossDistance
	leveragedPositionSize := positionSize * leverage
	marginRequired := leveragedPositionSize / leverage

	// Display trade information
	fmt.Printf("\nTrade Information:\n")
	fmt.Printf("Available margin: $%.2f\n", availableMargin)
	fmt.Printf("Selected pair: %s\n", selectedPair)
	fmt.Printf("Entry price: $%.2f\n", entryPrice)
	fmt.Printf("Stop loss price: $%.2f\n", stopLossPrice)
	fmt.Printf("Leverage: %.2fx\n", leverage)
	fmt.Printf("Maximum risk amount: $%.2f (%.2f%% of available margin)\n", maxRiskAmount, riskPercentage*100)
	fmt.Printf("Stop loss distance: %.2f%%\n", stopLossDistance * 100)
	fmt.Printf("Position size: $%.2f\n", positionSize)
	fmt.Printf("Leveraged position size: $%.2f\n", leveragedPositionSize)
	fmt.Printf("Margin required: $%.2f\n", marginRequired)

	if marginRequired > availableMargin {
		return fmt.Errorf("required margin ($%.2f) exceeds available margin ($%.2f)", marginRequired, availableMargin)
	}

	fmt.Printf("Final position size: $%.2f\n", leveragedPositionSize)
	fmt.Printf("Margin used: $%.2f\n", marginRequired)
	fmt.Printf("Effective leverage: %.2fx\n", leveragedPositionSize / marginRequired)

	// Confirm trade
	confirmQuestions := []ui.Question{
		ui.NewShortQuestion("Do you want to place this trade? (yes/no):"),
	}
	confirmWidget := ui.New(confirmQuestions)
	confirmP := tea.NewProgram(confirmWidget, tea.WithAltScreen())
	confirmM, err := confirmP.Run()
	if err != nil {
		return err
	}
	confirmAnswer := confirmM.(*ui.Main).GetAnswers()[0]

	if strings.ToLower(confirmAnswer) == "yes" {
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
