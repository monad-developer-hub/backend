package repository

import (
	"monad-devhub-be/internal/models"

	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

// CreateSubmission creates a new project submission
func (r *SubmissionRepository) CreateSubmission(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

// GetSubmissionByID retrieves a submission by ID
func (r *SubmissionRepository) GetSubmissionByID(submissionID string) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Preload("ApprovedProject.TeamMembers").First(&submission, "id = ?", submissionID).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetSubmissions retrieves submissions with pagination and filtering
func (r *SubmissionRepository) GetSubmissions(offset, limit int, status, sortBy, sortOrder string) ([]models.Submission, error) {
	query := r.db.Preload("ApprovedProject")

	// Apply status filter
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Apply sorting
	if sortBy != "" && sortOrder != "" {
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("submitted_at DESC")
	}

	var submissions []models.Submission
	err := query.Offset(offset).Limit(limit).Find(&submissions).Error
	return submissions, err
}

// GetSubmissionsCount returns total count with filters
func (r *SubmissionRepository) GetSubmissionsCount(status string) (int64, error) {
	query := r.db.Model(&models.Submission{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// UpdateSubmission updates an existing submission
func (r *SubmissionRepository) UpdateSubmission(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

// GetSubmissionByProjectName retrieves a submission by project name
func (r *SubmissionRepository) GetSubmissionByProjectName(projectName string) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Where("project_name = ?", projectName).First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetSubmissionStats returns statistics about submissions
func (r *SubmissionRepository) GetSubmissionStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	statuses := []string{"pending", "under_review", "approved", "rejected", "requires_changes"}
	for _, status := range statuses {
		var count int64
		err := r.db.Model(&models.Submission{}).Where("status = ?", status).Count(&count).Error
		if err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, nil
}
