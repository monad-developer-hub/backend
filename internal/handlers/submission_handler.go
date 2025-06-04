package handlers

import (
	"net/http"

	"monad-devhub-be/internal/services"
	"monad-devhub-be/internal/utils"

	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	projectService *services.ProjectService
}

func NewSubmissionHandler(projectService *services.ProjectService) *SubmissionHandler {
	return &SubmissionHandler{
		projectService: projectService,
	}
}

// SubmitProject handles POST /api/v1/submissions
// This is where the submission ID generation happens
func (h *SubmissionHandler) SubmitProject(c *gin.Context) {
	var req services.SubmitProjectRequest

	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SUBMISSION_DATA",
				"message": "Invalid submission data",
				"details": err.Error(),
			},
		})
		return
	}

	// Submit project through service (this generates the submission ID)
	response, err := h.projectService.SubmitProject(&req)
	if err != nil {
		// Parse error to determine appropriate status code
		if err.Error() == "DUPLICATE_PROJECT_NAME: Project with this name already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DUPLICATE_PROJECT_NAME",
					"message": "A project with this name already exists",
				},
			})
			return
		}

		if err.Error() == "DUPLICATE_SUBMISSION: Submission with this project name already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DUPLICATE_SUBMISSION",
					"message": "A submission with this project name already exists",
				},
			})
			return
		}

		// Handle validation errors
		if err.Error() == "INVALID_CATEGORIES: Invalid categories provided" ||
			err.Error() == "INVALID_EVENT: Invalid event provided" ||
			err.Error() == "INVALID_TEAM_MEMBERS: All team members must have name and twitter" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": err.Error(),
				},
			})
			return
		}

		// Generic error
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "SUBMISSION_FAILED",
				"message": "Failed to submit project",
				"details": err.Error(),
			},
		})
		return
	}

	// Return success response with submission ID
	c.JSON(http.StatusCreated, response)
}

// GetSubmissionStatus handles GET /api/v1/submissions/:submissionId
// This tracks the submission status using the submission ID
func (h *SubmissionHandler) GetSubmissionStatus(c *gin.Context) {
	submissionID := c.Param("submissionId")

	// Validate submission ID format
	if !utils.ValidateSubmissionID(submissionID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SUBMISSION_ID",
				"message": "Invalid submission ID format. Expected format: SUB-{timestamp}-{hash}",
			},
		})
		return
	}

	// Get submission status (implementation would be in a submission service)
	// For now, returning a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"submissionId": submissionID,
		"status":       "pending",
		"projectName":  "Sample Project",
		"submittedAt":  "2024-01-15T10:30:00Z",
		"timeline": gin.H{
			"submitted": "2024-01-15T10:30:00Z",
		},
	})
}

// GetSubmissions handles GET /api/v1/submissions (Admin only)
func (h *SubmissionHandler) GetSubmissions(c *gin.Context) {
	// TODO: Add authentication middleware for admin access

	// Get query parameters
	_ = c.Query("status") // TODO: Use this parameter for filtering

	// Placeholder response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"submissions": []gin.H{
			{
				"submissionId": "SUB-1749035470531-4W6UZJ",
				"status":       "pending",
				"projectName":  "Sample Project 1",
				"submittedAt":  "2024-01-15T10:30:00Z",
			},
			{
				"submissionId": "SUB-1749035471532-5X7VYK",
				"status":       "approved",
				"projectName":  "Sample Project 2",
				"submittedAt":  "2024-01-14T09:20:00Z",
				"reviewedAt":   "2024-01-16T14:30:00Z",
			},
		},
		"pagination": gin.H{
			"page":       1,
			"limit":      10,
			"total":      2,
			"totalPages": 1,
		},
		"stats": gin.H{
			"pending":          5,
			"under_review":     3,
			"approved":         12,
			"rejected":         2,
			"requires_changes": 1,
		},
	})
}

// ReviewSubmission handles PUT /api/v1/submissions/:submissionId/review (Admin only)
func (h *SubmissionHandler) ReviewSubmission(c *gin.Context) {
	submissionID := c.Param("submissionId")

	// Validate submission ID format
	if !utils.ValidateSubmissionID(submissionID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SUBMISSION_ID",
				"message": "Invalid submission ID format",
			},
		})
		return
	}

	// Parse review request
	var reviewRequest struct {
		Status           string   `json:"status" binding:"required,oneof=approved rejected requires_changes"`
		Feedback         *string  `json:"feedback,omitempty"`
		ChangesRequested []string `json:"changesRequested,omitempty"`
		ReviewerID       uint     `json:"reviewerId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reviewRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REVIEW_DATA",
				"message": "Invalid review data",
				"details": err.Error(),
			},
		})
		return
	}

	// TODO: Implement actual review logic

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"submissionId": submissionID,
		"newStatus":    reviewRequest.Status,
		"message":      "Submission reviewed successfully",
	})
}
