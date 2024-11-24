package main

import (
	"log"
	"github.com/sub0xdai/n0xtilus/internal/config"
	"github.com/sub0xdai/n0xtilus/internal/api"
	"github.com/sub0xdai/n0xtilus/internal/services"
	"github.com/sub0xdai/n0xtilus/internal/services/risk_calculator"
	"github.com/sub0xdai/n0xtilus/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type mainModel struct {
	dashboard    *ui.PositionDashboard
	tradeWidget  *ui.TradeInputWidget
	client       *api.APIClient
	orderService *services.OrderService
	riskPercent  float64
}

func (m mainModel) Init() tea.Cmd {
	return m.dashboard.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.(type) {
	case ui.ExecuteTradeMsg:
		m.tradeWidget = ui.NewTradeInputWidget([]string{"BTC/USD", "ETH/USD"}) // TODO: Get from API
		return m, m.tradeWidget.Init()
	}

	// Handle updates based on current active component
	if m.tradeWidget != nil {
		var cmd tea.Cmd
		updatedModel, cmd := m.tradeWidget.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
		if widget, ok := updatedModel.(*ui.TradeInputWidget); ok {
			m.tradeWidget = widget
			if widget.IsComplete() {
				if widget.Confirmed {
					// Process trade
					pairIdx, entry, stop, leverage, err := widget.GetInputs()
					if err != nil {
						log.Printf("Error getting inputs: %v", err)
					} else {
						// TODO: Execute trade
						log.Printf("Trade executed: pair=%d entry=%.2f stop=%.2f leverage=%.2f", 
							pairIdx, entry, stop, leverage)
					}
				}
				m.tradeWidget = nil
				return m, nil
			}
		}
	}

	// Update dashboard
	var dashCmd tea.Cmd
	updatedModel, dashCmd := m.dashboard.Update(msg)
	if dashCmd != nil {
		cmds = append(cmds, dashCmd)
	}

	// Type assert the updated dashboard model
	if updatedDashboard, ok := updatedModel.(*ui.PositionDashboard); ok {
		m.dashboard = updatedDashboard
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	if m.tradeWidget != nil {
		return m.tradeWidget.View()
	}
	return m.dashboard.View()
}

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

	// Create and run the main application
	model := mainModel{
		dashboard:    ui.NewPositionDashboard(true), // Using placeholder data for now
		client:      client,
		orderService: orderService,
		riskPercent: cfg.RiskPercentage / 100,
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
