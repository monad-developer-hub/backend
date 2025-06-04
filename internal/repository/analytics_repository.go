package repository

import (
	"time"

	"monad-devhub-be/internal/models"

	"gorm.io/gorm"
)

type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetLatestStats retrieves the most recent analytics stats
func (r *AnalyticsRepository) GetLatestStats() (*models.AnalyticsStats, error) {
	var stats models.AnalyticsStats
	err := r.db.Order("timestamp DESC").First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// CreateStats creates new analytics stats entry
func (r *AnalyticsRepository) CreateStats(stats *models.AnalyticsStats) error {
	return r.db.Create(stats).Error
}

// GetTransactions retrieves transactions with pagination and filtering
func (r *AnalyticsRepository) GetTransactions(offset, limit int, txType string, from, to *time.Time) ([]models.Transaction, error) {
	query := r.db.Model(&models.Transaction{})

	// Apply filters
	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if from != nil {
		query = query.Where("timestamp >= ?", from)
	}
	if to != nil {
		query = query.Where("timestamp <= ?", to)
	}

	var transactions []models.Transaction
	err := query.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	return transactions, err
}

// GetTransactionsCount returns total count with filters
func (r *AnalyticsRepository) GetTransactionsCount(txType string, from, to *time.Time) (int64, error) {
	query := r.db.Model(&models.Transaction{})

	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if from != nil {
		query = query.Where("timestamp >= ?", from)
	}
	if to != nil {
		query = query.Where("timestamp <= ?", to)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// CreateTransaction creates a new transaction record
func (r *AnalyticsRepository) CreateTransaction(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

// GetTopContracts retrieves top contracts by activity
func (r *AnalyticsRepository) GetTopContracts(limit int, period string, sortBy string) ([]models.ContractStats, error) {
	query := r.db.Preload("Contract")

	// Apply time filter based on period
	if period != "" {
		var since time.Time
		switch period {
		case "1h":
			since = time.Now().Add(-time.Hour)
		case "24h":
			since = time.Now().Add(-24 * time.Hour)
		case "7d":
			since = time.Now().Add(-7 * 24 * time.Hour)
		case "30d":
			since = time.Now().Add(-30 * 24 * time.Hour)
		}
		if !since.IsZero() {
			query = query.Where("last_updated >= ?", since)
		}
	}

	// Apply sorting
	switch sortBy {
	case "uniqueWallets":
		query = query.Order("unique_wallets DESC")
	case "gasUsed":
		query = query.Order("gas_used DESC")
	default:
		query = query.Order("tx_count DESC")
	}

	var contractStats []models.ContractStats
	err := query.Limit(limit).Find(&contractStats).Error
	return contractStats, err
}

// UpdateContractStats updates contract statistics
func (r *AnalyticsRepository) UpdateContractStats(stats *models.ContractStats) error {
	return r.db.Save(stats).Error
}

// GetContract retrieves a contract by address
func (r *AnalyticsRepository) GetContract(address string) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.Where("address = ?", address).First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

// CreateContract creates a new contract
func (r *AnalyticsRepository) CreateContract(contract *models.Contract) error {
	return r.db.Create(contract).Error
}
