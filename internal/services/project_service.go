package services

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"monad-devhub-be/internal/models"
	"monad-devhub-be/internal/repository"
	"monad-devhub-be/internal/utils"

	"gorm.io/gorm"
)

type ProjectService struct {
	projectRepo    *repository.ProjectRepository
	submissionRepo *repository.SubmissionRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository, submissionRepo *repository.SubmissionRepository) *ProjectService {
	return &ProjectService{
		projectRepo:    projectRepo,
		submissionRepo: submissionRepo,
	}
}

// SubmitProjectRequest represents the request payload for project submission
type SubmitProjectRequest struct {
	PhotoLink       string                   `json:"photoLink" binding:"required"`
	ProjectName     string                   `json:"projectName" binding:"required"`
	Description     string                   `json:"description" binding:"required"`
	Event           string                   `json:"event" binding:"required"`
	Categories      []string                 `json:"categories" binding:"required,min=1"`
	TeamMembers     []models.TeamMemberInput `json:"teamMembers" binding:"required,min=1"`
	GithubLink      *string                  `json:"githubLink,omitempty"`
	WebsiteLink     *string                  `json:"websiteLink,omitempty"`
	PlayLink        string                   `json:"playLink" binding:"required"`
	HowToPlay       string                   `json:"howToPlay" binding:"required"`
	AdditionalNotes *string                  `json:"additionalNotes,omitempty"`
}

// SubmitProjectResponse represents the response for project submission
type SubmitProjectResponse struct {
	Success             bool     `json:"success"`
	SubmissionID        string   `json:"submissionId"`
	Message             string   `json:"message"`
	EstimatedReviewTime string   `json:"estimatedReviewTime"`
	NextSteps           []string `json:"nextSteps"`
}

// GetProjectsRequest represents the request for getting projects
type GetProjectsRequest struct {
	Page      int      `form:"page" binding:"min=1"`
	Limit     int      `form:"limit" binding:"min=1,max=100"`
	Category  []string `form:"category"`
	Event     string   `form:"event"`
	Award     string   `form:"award"`
	Search    string   `form:"search"`
	SortBy    string   `form:"sortBy"`
	SortOrder string   `form:"sortOrder"`
}

// GetProjectsResponse represents the response for getting projects
type GetProjectsResponse struct {
	Projects   []models.Project `json:"projects"`
	Pagination PaginationInfo   `json:"pagination"`
	Filters    FilterInfo       `json:"filters"`
}

type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type FilterInfo struct {
	Categories []string `json:"categories"`
	Events     []string `json:"events"`
	Awards     []string `json:"awards"`
}

// SubmitProject handles project submission with validation and submission ID generation
func (s *ProjectService) SubmitProject(req *SubmitProjectRequest) (*SubmitProjectResponse, error) {
	// Validate request
	if err := s.validateSubmissionRequest(req); err != nil {
		return nil, err
	}

	// Check for duplicate project name
	_, err := s.projectRepo.GetProjectByName(req.ProjectName)
	if err == nil {
		return nil, errors.New("DUPLICATE_PROJECT_NAME: Project with this name already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Check for duplicate submission
	_, err = s.submissionRepo.GetSubmissionByProjectName(req.ProjectName)
	if err == nil {
		return nil, errors.New("DUPLICATE_SUBMISSION: Submission with this project name already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Generate unique submission ID
	submissionID := utils.GenerateSubmissionID()

	// Convert team members to JSON
	teamMembersJSON, err := json.Marshal(req.TeamMembers)
	if err != nil {
		return nil, err
	}

	// Create submission
	submission := &models.Submission{
		ID:              submissionID,
		ProjectName:     req.ProjectName,
		Description:     req.Description,
		PhotoLink:       req.PhotoLink,
		Event:           req.Event,
		Categories:      req.Categories,
		TeamMembers:     string(teamMembersJSON),
		GithubLink:      req.GithubLink,
		WebsiteLink:     req.WebsiteLink,
		PlayLink:        req.PlayLink,
		HowToPlay:       req.HowToPlay,
		AdditionalNotes: req.AdditionalNotes,
		Status:          "pending",
		SubmittedAt:     time.Now(),
	}

	if err := s.submissionRepo.CreateSubmission(submission); err != nil {
		return nil, err
	}

	// Return success response
	return &SubmitProjectResponse{
		Success:             true,
		SubmissionID:        submissionID,
		Message:             "Your project has been submitted successfully!",
		EstimatedReviewTime: "2-3 business days",
		NextSteps: []string{
			"We'll review your submission within 2-3 business days",
			"You'll receive an email update when review is complete",
			"Use submission ID " + submissionID + " to check status anytime",
		},
	}, nil
}

// GetProjects retrieves projects with pagination and filtering
func (s *ProjectService) GetProjects(req *GetProjectsRequest) (*GetProjectsResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	// Get projects
	projects, err := s.projectRepo.GetProjects(
		offset, req.Limit, req.Category, req.Event, req.Award,
		req.Search, req.SortBy, req.SortOrder,
	)
	if err != nil {
		return nil, err
	}

	// Get total count
	total, err := s.projectRepo.GetProjectsCount(req.Category, req.Event, req.Award, req.Search)
	if err != nil {
		return nil, err
	}

	// Get filter options
	categories, _ := s.projectRepo.GetDistinctCategories()
	events, _ := s.projectRepo.GetDistinctEvents()
	awards, _ := s.projectRepo.GetDistinctAwards()

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &GetProjectsResponse{
		Projects: projects,
		Pagination: PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
		Filters: FilterInfo{
			Categories: categories,
			Events:     events,
			Awards:     awards,
		},
	}, nil
}

// GetProject retrieves a single project by ID
func (s *ProjectService) GetProject(id uint) (*models.Project, error) {
	return s.projectRepo.GetProjectByID(id)
}

// LikeProject increments the likes count for a project
func (s *ProjectService) LikeProject(id uint) error {
	// Check if project exists
	_, err := s.projectRepo.GetProjectByID(id)
	if err != nil {
		return err
	}

	return s.projectRepo.IncrementLikes(id)
}

// validateSubmissionRequest validates the submission request
func (s *ProjectService) validateSubmissionRequest(req *SubmitProjectRequest) error {
	// Validate categories
	if !utils.ValidateCategories(req.Categories) {
		return errors.New("INVALID_CATEGORIES: Invalid categories provided")
	}

	// Validate event
	if !utils.ValidateEvent(req.Event) {
		return errors.New("INVALID_EVENT: Invalid event provided")
	}

	// Validate team members
	for _, member := range req.TeamMembers {
		if member.Name == "" || member.Twitter == "" {
			return errors.New("INVALID_TEAM_MEMBERS: All team members must have name and twitter")
		}
	}

	return nil
}
