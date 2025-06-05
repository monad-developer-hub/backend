package handlers

import (
	"net/http"
	"strconv"

	"monad-devhub-be/internal/services"
	"monad-devhub-be/internal/utils"

	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	projectService    *services.ProjectService
	submissionService *services.SubmissionService
}

func NewSubmissionHandler(projectService *services.ProjectService, submissionService *services.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{
		projectService:    projectService,
		submissionService: submissionService,
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

	// Get submission from service
	submission, err := h.submissionService.GetSubmissionByID(submissionID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SUBMISSION_NOT_FOUND",
					"message": "Submission not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve submission",
				"details": err.Error(),
			},
		})
		return
	}

	// Build timeline
	timeline := gin.H{
		"submitted": submission.SubmittedAt,
	}
	if submission.ReviewStartedAt != nil {
		timeline["review_started"] = *submission.ReviewStartedAt
	}
	if submission.ReviewedAt != nil {
		timeline["review_completed"] = *submission.ReviewedAt
	}
	if submission.PublishedAt != nil {
		timeline["published"] = *submission.PublishedAt
	}

	// Return submission status
	c.JSON(http.StatusOK, gin.H{
		"submissionId": submission.ID,
		"status":       submission.Status,
		"projectName":  submission.ProjectName,
		"submittedAt":  submission.SubmittedAt,
		"reviewedAt":   submission.ReviewedAt,
		"feedback":     submission.Feedback,
		"project":      submission.ApprovedProject,
		"timeline":     timeline,
	})
}

// GetSubmissions handles GET /api/v1/submissions
func (h *SubmissionHandler) GetSubmissions(c *gin.Context) {

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	sortBy := c.DefaultQuery("sortBy", "submitted_at")
	sortOrder := c.DefaultQuery("sortOrder", "DESC")

	// Create request
	req := &services.GetSubmissionsRequest{
		Page:      page,
		Limit:     limit,
		Status:    status,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	// Get submissions from service
	response, err := h.submissionService.GetSubmissions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve submissions",
				"details": err.Error(),
			},
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// ReviewSubmission handles PUT /api/v1/submissions/:submissionId/review
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
		Status           string   `json:"status" binding:"required,oneof=pending under_review approved rejected requires_changes"`
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

	// Update submission status
	err := h.submissionService.UpdateSubmissionStatus(
		submissionID,
		reviewRequest.Status,
		reviewRequest.Feedback,
		reviewRequest.ChangesRequested,
		&reviewRequest.ReviewerID,
	)

	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SUBMISSION_NOT_FOUND",
					"message": "Submission not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update submission",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"submissionId": submissionID,
		"newStatus":    reviewRequest.Status,
		"message":      "Submission reviewed successfully",
	})
}

// UpdateProjectExtras handles PUT /api/v1/submissions/:submissionId/project-extras
// Admin-only endpoint to update project award and team member photos after review
func (h *SubmissionHandler) UpdateProjectExtras(c *gin.Context) {
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

	// Parse update request
	var updateRequest struct {
		Award      *string             `json:"award,omitempty"`
		TeamPhotos []map[string]string `json:"teamPhotos,omitempty"` // [{memberName: "John", photoUrl: "https://..."}]
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_UPDATE_DATA",
				"message": "Invalid update data",
				"details": err.Error(),
			},
		})
		return
	}

	// Update project extras
	err := h.submissionService.UpdateProjectExtras(
		submissionID,
		updateRequest.Award,
		updateRequest.TeamPhotos,
	)

	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SUBMISSION_NOT_FOUND",
					"message": "Submission not found",
				},
			})
			return
		}

		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "PROJECT_NOT_FOUND",
					"message": "Project not found - submission may not be approved yet",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update project extras",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"submissionId": submissionID,
		"message":      "Project extras updated successfully",
	})
}
