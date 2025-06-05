package services

import (
	"encoding/json"
	"math"
	"time"

	"errors"
	"monad-devhub-be/internal/models"
	"monad-devhub-be/internal/repository"
)

type SubmissionService struct {
	submissionRepo *repository.SubmissionRepository
	projectRepo    *repository.ProjectRepository
}

func NewSubmissionService(submissionRepo *repository.SubmissionRepository, projectRepo *repository.ProjectRepository) *SubmissionService {
	return &SubmissionService{
		submissionRepo: submissionRepo,
		projectRepo:    projectRepo,
	}
}

// GetSubmissionsRequest represents the request for getting submissions
type GetSubmissionsRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	Limit     int    `form:"limit" binding:"min=1,max=100"`
	Status    string `form:"status"`
	SortBy    string `form:"sortBy"`
	SortOrder string `form:"sortOrder"`
}

// GetSubmissionsResponse represents the response for getting submissions
type GetSubmissionsResponse struct {
	Success     bool                        `json:"success"`
	Submissions []SubmissionWithTeamMembers `json:"submissions"`
	Pagination  PaginationInfo              `json:"pagination"`
	Stats       map[string]int64            `json:"stats"`
}

// SubmissionWithTeamMembers represents a submission with parsed team members
type SubmissionWithTeamMembers struct {
	ID                string                   `json:"submissionId"`
	ProjectName       string                   `json:"projectName"`
	Description       string                   `json:"description"`
	PhotoLink         string                   `json:"photoLink"`
	Event             string                   `json:"event"`
	Categories        []string                 `json:"categories"`
	TeamMembers       []models.TeamMemberInput `json:"teamMembers"`
	GithubLink        *string                  `json:"githubLink,omitempty"`
	WebsiteLink       *string                  `json:"websiteLink,omitempty"`
	PlayLink          string                   `json:"playLink"`
	HowToPlay         string                   `json:"howToPlay"`
	AdditionalNotes   *string                  `json:"additionalNotes,omitempty"`
	Status            string                   `json:"status"`
	ReviewerID        *uint                    `json:"reviewerId,omitempty"`
	Feedback          *string                  `json:"feedback,omitempty"`
	ChangesRequested  []string                 `json:"changesRequested,omitempty"`
	SubmittedAt       string                   `json:"submittedAt"`
	ReviewStartedAt   *string                  `json:"reviewStartedAt,omitempty"`
	ReviewedAt        *string                  `json:"reviewedAt,omitempty"`
	PublishedAt       *string                  `json:"publishedAt,omitempty"`
	ApprovedProjectID *uint                    `json:"approvedProjectId,omitempty"`
	ApprovedProject   *models.Project          `json:"project,omitempty"`
}

// GetSubmissions retrieves submissions with pagination and filtering
func (s *SubmissionService) GetSubmissions(req *GetSubmissionsRequest) (*GetSubmissionsResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	// Get submissions from repository
	submissions, err := s.submissionRepo.GetSubmissions(offset, req.Limit, req.Status, req.SortBy, req.SortOrder)
	if err != nil {
		return nil, err
	}

	// Get total count
	total, err := s.submissionRepo.GetSubmissionsCount(req.Status)
	if err != nil {
		return nil, err
	}

	// Get statistics
	stats, err := s.submissionRepo.GetSubmissionStats()
	if err != nil {
		return nil, err
	}

	// Convert submissions to response format with parsed team members
	var submissionResponses []SubmissionWithTeamMembers
	for _, submission := range submissions {
		// Parse team members JSON
		var teamMembers []models.TeamMemberInput
		if submission.TeamMembers != "" {
			json.Unmarshal([]byte(submission.TeamMembers), &teamMembers)
		}

		submissionResponse := SubmissionWithTeamMembers{
			ID:                submission.ID,
			ProjectName:       submission.ProjectName,
			Description:       submission.Description,
			PhotoLink:         submission.PhotoLink,
			Event:             submission.Event,
			Categories:        submission.Categories,
			TeamMembers:       teamMembers,
			GithubLink:        submission.GithubLink,
			WebsiteLink:       submission.WebsiteLink,
			PlayLink:          submission.PlayLink,
			HowToPlay:         submission.HowToPlay,
			AdditionalNotes:   submission.AdditionalNotes,
			Status:            submission.Status,
			ReviewerID:        submission.ReviewerID,
			Feedback:          submission.Feedback,
			ChangesRequested:  submission.ChangesRequested,
			SubmittedAt:       submission.SubmittedAt.Format("2006-01-02T15:04:05Z"),
			ApprovedProjectID: submission.ApprovedProjectID,
			ApprovedProject:   submission.ApprovedProject,
		}

		// Format optional timestamps
		if submission.ReviewStartedAt != nil {
			timestamp := submission.ReviewStartedAt.Format("2006-01-02T15:04:05Z")
			submissionResponse.ReviewStartedAt = &timestamp
		}
		if submission.ReviewedAt != nil {
			timestamp := submission.ReviewedAt.Format("2006-01-02T15:04:05Z")
			submissionResponse.ReviewedAt = &timestamp
		}
		if submission.PublishedAt != nil {
			timestamp := submission.PublishedAt.Format("2006-01-02T15:04:05Z")
			submissionResponse.PublishedAt = &timestamp
		}

		submissionResponses = append(submissionResponses, submissionResponse)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &GetSubmissionsResponse{
		Success:     true,
		Submissions: submissionResponses,
		Pagination: PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
		Stats: stats,
	}, nil
}

// GetSubmissionByID retrieves a single submission by ID
func (s *SubmissionService) GetSubmissionByID(submissionID string) (*SubmissionWithTeamMembers, error) {
	submission, err := s.submissionRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}

	// Parse team members JSON
	var teamMembers []models.TeamMemberInput
	if submission.TeamMembers != "" {
		json.Unmarshal([]byte(submission.TeamMembers), &teamMembers)
	}

	submissionResponse := &SubmissionWithTeamMembers{
		ID:                submission.ID,
		ProjectName:       submission.ProjectName,
		Description:       submission.Description,
		PhotoLink:         submission.PhotoLink,
		Event:             submission.Event,
		Categories:        submission.Categories,
		TeamMembers:       teamMembers,
		GithubLink:        submission.GithubLink,
		WebsiteLink:       submission.WebsiteLink,
		PlayLink:          submission.PlayLink,
		HowToPlay:         submission.HowToPlay,
		AdditionalNotes:   submission.AdditionalNotes,
		Status:            submission.Status,
		ReviewerID:        submission.ReviewerID,
		Feedback:          submission.Feedback,
		ChangesRequested:  submission.ChangesRequested,
		SubmittedAt:       submission.SubmittedAt.Format("2006-01-02T15:04:05Z"),
		ApprovedProjectID: submission.ApprovedProjectID,
		ApprovedProject:   submission.ApprovedProject,
	}

	// Format optional timestamps
	if submission.ReviewStartedAt != nil {
		timestamp := submission.ReviewStartedAt.Format("2006-01-02T15:04:05Z")
		submissionResponse.ReviewStartedAt = &timestamp
	}
	if submission.ReviewedAt != nil {
		timestamp := submission.ReviewedAt.Format("2006-01-02T15:04:05Z")
		submissionResponse.ReviewedAt = &timestamp
	}
	if submission.PublishedAt != nil {
		timestamp := submission.PublishedAt.Format("2006-01-02T15:04:05Z")
		submissionResponse.PublishedAt = &timestamp
	}

	return submissionResponse, nil
}

// UpdateSubmissionStatus updates the status of a submission
func (s *SubmissionService) UpdateSubmissionStatus(submissionID string, status string, feedback *string, changesRequested []string, reviewerID *uint) error {
	submission, err := s.submissionRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return err
	}

	// Store previous status to check if this is a new approval
	previousStatus := submission.Status

	// Update fields
	submission.Status = status
	submission.Feedback = feedback
	submission.ChangesRequested = changesRequested
	submission.ReviewerID = reviewerID

	// Set timestamps based on status
	now := time.Now()
	switch status {
	case "under_review":
		if submission.ReviewStartedAt == nil {
			submission.ReviewStartedAt = &now
		}
	case "approved", "rejected", "requires_changes":
		submission.ReviewedAt = &now
		// If being approved for the first time and no project exists yet
		if status == "approved" && previousStatus != "approved" && submission.ApprovedProjectID == nil {
			if err := s.createProjectFromSubmission(submission); err != nil {
				return err
			}
			submission.PublishedAt = &now
		}
	}

	return s.submissionRepo.UpdateSubmission(submission)
}

// createProjectFromSubmission creates a new project from an approved submission
func (s *SubmissionService) createProjectFromSubmission(submission *models.Submission) error {
	// Parse team members from JSON
	var teamMembersInput []models.TeamMemberInput
	if submission.TeamMembers != "" {
		if err := json.Unmarshal([]byte(submission.TeamMembers), &teamMembersInput); err != nil {
			return err
		}
	}

	// Create the project
	project := &models.Project{
		Name:         submission.ProjectName,
		Logo:         submission.PhotoLink,
		Description:  submission.Description,
		Categories:   submission.Categories,
		Event:        submission.Event,
		Award:        "", // Can be set later
		Likes:        0,
		Comments:     0,
		HowToPlay:    submission.HowToPlay,
		PlayURL:      submission.PlayLink,
		GithubURL:    submission.GithubLink,
		WebsiteURL:   submission.WebsiteLink,
		SubmissionID: &submission.ID,
	}

	// Create the project in database
	if err := s.projectRepo.CreateProject(project); err != nil {
		return err
	}

	// Create team members for the project
	var teamMembers []models.TeamMember
	for _, memberInput := range teamMembersInput {
		teamMember := models.TeamMember{
			ProjectID: project.ID,
			Name:      memberInput.Name,
			Twitter:   memberInput.Twitter,
			Image:     "", // Will be empty initially, can be updated later
		}
		teamMembers = append(teamMembers, teamMember)
	}

	// Set team members to the project and update it
	if len(teamMembers) > 0 {
		project.TeamMembers = teamMembers
		if err := s.projectRepo.UpdateProject(project); err != nil {
			return err
		}
	}

	// Link the submission to the created project
	submission.ApprovedProjectID = &project.ID

	return nil
}

// UpdateProjectExtras updates project award and team member photos for an approved submission
func (s *SubmissionService) UpdateProjectExtras(submissionID string, award *string, teamPhotos []map[string]string) error {
	// Get submission
	submission, err := s.submissionRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return err
	}

	// Check if submission has an approved project
	if submission.ApprovedProjectID == nil {
		return errors.New("project not found")
	}

	// Get the project
	project, err := s.projectRepo.GetProjectByID(*submission.ApprovedProjectID)
	if err != nil {
		return errors.New("project not found")
	}

	// Update award if provided
	if award != nil {
		project.Award = *award
	}

	// Update team member photos if provided
	if len(teamPhotos) > 0 {
		// Create a map of member names to photo URLs for quick lookup
		photoMap := make(map[string]string)
		for _, teamPhoto := range teamPhotos {
			if memberName, ok := teamPhoto["memberName"]; ok {
				if photoUrl, ok := teamPhoto["photoUrl"]; ok && photoUrl != "" {
					photoMap[memberName] = photoUrl
				}
			}
		}

		// Update team member photos
		for i := range project.TeamMembers {
			if photoUrl, exists := photoMap[project.TeamMembers[i].Name]; exists {
				project.TeamMembers[i].Image = photoUrl
			}
		}
	}

	// Save the updated project
	return s.projectRepo.UpdateProject(project)
}
