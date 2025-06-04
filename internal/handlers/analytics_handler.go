package handlers

import (
	"net/http"

	"monad-devhub-be/internal/services"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetStats handles GET /api/v1/analytics/stats
func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	response, err := h.analyticsService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ANALYTICS_SERVICE_UNAVAILABLE",
				"message": "Failed to retrieve analytics stats",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactions handles GET /api/v1/analytics/transactions
func (h *AnalyticsHandler) GetTransactions(c *gin.Context) {
	var req services.GetTransactionsRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid query parameters",
				"details": err.Error(),
			},
		})
		return
	}

	// Get transactions from service
	response, err := h.analyticsService.GetTransactions(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ANALYTICS_SERVICE_UNAVAILABLE",
				"message": "Failed to retrieve transactions",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTopContracts handles GET /api/v1/analytics/contracts/top
func (h *AnalyticsHandler) GetTopContracts(c *gin.Context) {
	var req services.GetTopContractsRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid query parameters",
				"details": err.Error(),
			},
		})
		return
	}

	// Get top contracts from service
	response, err := h.analyticsService.GetTopContracts(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ANALYTICS_SERVICE_UNAVAILABLE",
				"message": "Failed to retrieve top contracts",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
