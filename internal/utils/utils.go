package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GenerateSubmissionID generates a unique submission ID in the format SUB-{timestamp}-{randomHash}
func GenerateSubmissionID() string {
	timestamp := time.Now().UnixMilli() // 1749035470531
	randomHash := generateRandomHash(6) // 4W6UZJ
	return fmt.Sprintf("SUB-%d-%s", timestamp, randomHash)
}

// generateRandomHash generates a random alphanumeric hash of specified length
func generateRandomHash(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	var result strings.Builder
	for i := 0; i < length; i++ {
		result.WriteByte(charset[rand.Intn(len(charset))])
	}
	return result.String()
}

// ValidateSubmissionID validates the format of a submission ID
func ValidateSubmissionID(submissionID string) bool {
	parts := strings.Split(submissionID, "-")
	if len(parts) != 3 {
		return false
	}

	// Check prefix
	if parts[0] != "SUB" {
		return false
	}

	// Check timestamp part (should be numeric)
	if len(parts[1]) < 10 {
		return false
	}

	// Check hash part (should be 6 characters)
	if len(parts[2]) != 6 {
		return false
	}

	return true
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveEmpty removes empty strings from a slice
func RemoveEmpty(slice []string) []string {
	var result []string
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			result = append(result, strings.TrimSpace(s))
		}
	}
	return result
}

// ValidateCategories validates that all categories are in the allowed list
func ValidateCategories(categories []string) bool {
	allowedCategories := []string{
		"DeFi", "Gaming", "AI", "Infrastructure", "Consumer", "NFT", "Stablecoins",
	}

	for _, category := range categories {
		if !Contains(allowedCategories, category) {
			return false
		}
	}
	return true
}

// ValidateEvents validates that the event is in the allowed list
func ValidateEvent(event string) bool {
	allowedEvents := []string{
		"Mission: 1 Crazy Contract",
		"Mission: 2 Smart Wallet",
		"Mission: 3 DeFi Integration",
		"Mission: 4 NFT Marketplace",
		"Hackathon 2023",
		"Hackathon 2024",
	}

	return Contains(allowedEvents, event)
}

// ValidateStatus validates submission status
func ValidateStatus(status string) bool {
	allowedStatuses := []string{
		"pending", "under_review", "approved", "rejected", "requires_changes",
	}

	return Contains(allowedStatuses, status)
}

// ValidateTransactionType validates transaction type
func ValidateTransactionType(txType string) bool {
	allowedTypes := []string{
		"transfer", "swap", "mint", "burn", "stake",
	}

	return Contains(allowedTypes, txType)
}
