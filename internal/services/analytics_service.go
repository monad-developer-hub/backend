package services

import (
	"time"

	"monad-devhub-be/internal/models"
	"monad-devhub-be/internal/repository"
	"monad-devhub-be/internal/utils"
)

type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
	}
}

// GetStatsResponse represents the response for analytics stats
type GetStatsResponse struct {
	Stats   *models.AnalyticsStats `json:"stats"`
	Success bool                   `json:"success"`
}

// GetTransactionsRequest represents the request for getting transactions
type GetTransactionsRequest struct {
	Limit int     `form:"limit" binding:"max=100"`
	Type  string  `form:"type"`
	From  *string `form:"from"`
	To    *string `form:"to"`
}

// GetTransactionsResponse represents the response for getting transactions
type GetTransactionsResponse struct {
	Transactions []models.Transaction  `json:"transactions"`
	Total        int64                 `json:"total"`
	Pagination   TransactionPagination `json:"pagination"`
}

type TransactionPagination struct {
	HasMore    bool   `json:"hasMore"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// GetTopContractsRequest represents the request for getting top contracts
type GetTopContractsRequest struct {
	Period string `form:"period" binding:"oneof=1h 24h 7d 30d"`
	Limit  int    `form:"limit" binding:"max=50"`
	SortBy string `form:"sortBy" binding:"oneof=txCount uniqueWallets gasUsed"`
}

// GetTopContractsResponse represents the response for getting top contracts
type GetTopContractsResponse struct {
	Contracts   []models.ContractStats `json:"contracts"`
	LastUpdated time.Time              `json:"lastUpdated"`
}

// GetStats retrieves the latest blockchain analytics statistics
func (s *AnalyticsService) GetStats() (*GetStatsResponse, error) {
	stats, err := s.analyticsRepo.GetLatestStats()
	if err != nil {
		// Return mock data if no stats found (for demo purposes)
		mockStats := &models.AnalyticsStats{
			TotalTransactions: 1737085372,
			TPS:               0,
			ActiveValidators:  99,
			BlockHeight:       0,
			Timestamp:         time.Now(),
		}
		return &GetStatsResponse{
			Stats:   mockStats,
			Success: true,
		}, nil
	}

	return &GetStatsResponse{
		Stats:   stats,
		Success: true,
	}, nil
}

// GetTransactions retrieves transactions with filtering
func (s *AnalyticsService) GetTransactions(req *GetTransactionsRequest) (*GetTransactionsResponse, error) {
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 50
	}

	// Parse time filters
	var fromTime, toTime *time.Time
	if req.From != nil {
		if parsed, err := time.Parse(time.RFC3339, *req.From); err == nil {
			fromTime = &parsed
		}
	}
	if req.To != nil {
		if parsed, err := time.Parse(time.RFC3339, *req.To); err == nil {
			toTime = &parsed
		}
	}

	// Validate transaction type
	if req.Type != "" && !utils.ValidateTransactionType(req.Type) {
		req.Type = "" // Ignore invalid type
	}

	// Get transactions
	transactions, err := s.analyticsRepo.GetTransactions(0, req.Limit, req.Type, fromTime, toTime)
	if err != nil {
		return nil, err
	}

	// Get total count
	total, err := s.analyticsRepo.GetTransactionsCount(req.Type, fromTime, toTime)
	if err != nil {
		return nil, err
	}

	return &GetTransactionsResponse{
		Transactions: transactions,
		Total:        total,
		Pagination: TransactionPagination{
			HasMore: len(transactions) == req.Limit,
		},
	}, nil
}

// GetTopContracts retrieves top contracts by activity
func (s *AnalyticsService) GetTopContracts(req *GetTopContractsRequest) (*GetTopContractsResponse, error) {
	// Set defaults
	if req.Period == "" {
		req.Period = "24h"
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.SortBy == "" {
		req.SortBy = "txCount"
	}

	// Get top contracts
	contracts, err := s.analyticsRepo.GetTopContracts(req.Limit, req.Period, req.SortBy)
	if err != nil {
		return nil, err
	}

	lastUpdated := time.Now()
	if len(contracts) > 0 {
		lastUpdated = contracts[0].LastUpdated
	}

	return &GetTopContractsResponse{
		Contracts:   contracts,
		LastUpdated: lastUpdated,
	}, nil
}
