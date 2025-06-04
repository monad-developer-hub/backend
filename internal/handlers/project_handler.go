package handlers

import (
	"net/http"
	"strconv"

	"monad-devhub-be/internal/services"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService *services.ProjectService
}

func NewProjectHandler(projectService *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// GetProjects handles GET /api/v1/projects
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	var req services.GetProjectsRequest

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

	// Get projects from service
	response, err := h.projectService.GetProjects(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve projects",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateProject handles POST /api/v1/projects (for direct project creation)
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Direct project creation is not implemented. Use submissions endpoint instead.",
		},
	})
}

// GetProject handles GET /api/v1/projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	// Parse project ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PROJECT_ID",
				"message": "Invalid project ID format",
			},
		})
		return
	}

	// Get project from service
	project, err := h.projectService.GetProject(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PROJECT_NOT_FOUND",
				"message": "Project not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"project": project,
	})
}

// LikeProject handles POST /api/v1/projects/:id/like
func (h *ProjectHandler) LikeProject(c *gin.Context) {
	// Parse project ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PROJECT_ID",
				"message": "Invalid project ID format",
			},
		})
		return
	}

	// Like project
	err = h.projectService.LikeProject(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PROJECT_NOT_FOUND",
				"message": "Project not found",
			},
		})
		return
	}

	// Get updated project to return new likes count
	project, err := h.projectService.GetProject(uint(id))
	if err != nil {
		// Still return success even if we can't get updated count
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Project liked successfully",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"likes":   project.Likes,
		"message": "Project liked successfully",
	})
}
