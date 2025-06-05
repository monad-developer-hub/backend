package repository

import (
	"monad-devhub-be/internal/models"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// GetProjects retrieves projects with pagination and filtering
func (r *ProjectRepository) GetProjects(offset, limit int, categories []string, event, award, search, sortBy, sortOrder string) ([]models.Project, error) {
	query := r.db.Preload("TeamMembers")

	// Apply filters
	if len(categories) > 0 {
		query = query.Where("categories && ?", categories)
	}
	if event != "" {
		query = query.Where("event = ?", event)
	}
	if award != "" {
		query = query.Where("award = ?", award)
	}
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Apply sorting
	if sortBy != "" && sortOrder != "" {
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("created_at DESC")
	}

	var projects []models.Project
	err := query.Offset(offset).Limit(limit).Find(&projects).Error
	return projects, err
}

// GetProjectsCount returns total count with filters
func (r *ProjectRepository) GetProjectsCount(categories []string, event, award, search string) (int64, error) {
	query := r.db.Model(&models.Project{})

	// Apply same filters as GetProjects
	if len(categories) > 0 {
		query = query.Where("categories && ?", categories)
	}
	if event != "" {
		query = query.Where("event = ?", event)
	}
	if award != "" {
		query = query.Where("award = ?", award)
	}
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// GetProjectByID retrieves a project by ID with team members
func (r *ProjectRepository) GetProjectByID(id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("TeamMembers").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject creates a new project
func (r *ProjectRepository) CreateProject(project *models.Project) error {
	return r.db.Create(project).Error
}

// UpdateProject updates an existing project and its team members
func (r *ProjectRepository) UpdateProject(project *models.Project) error {
	// Use a transaction to ensure atomicity
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update the project itself (excluding team members to handle them separately)
		if err := tx.Omit("TeamMembers").Save(project).Error; err != nil {
			return err
		}

		// Update each team member individually to ensure associations are saved
		for _, member := range project.TeamMembers {
			if err := tx.Save(&member).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetProjectByName retrieves a project by name
func (r *ProjectRepository) GetProjectByName(name string) (*models.Project, error) {
	var project models.Project
	err := r.db.Where("name = ?", name).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// IncrementLikes increments the likes count for a project
func (r *ProjectRepository) IncrementLikes(id uint) error {
	return r.db.Model(&models.Project{}).Where("id = ?", id).UpdateColumn("likes", gorm.Expr("likes + ?", 1)).Error
}

// GetDistinctCategories returns all unique categories
func (r *ProjectRepository) GetDistinctCategories() ([]string, error) {
	var categories []string
	err := r.db.Model(&models.Project{}).Distinct("unnest(categories)").Pluck("unnest", &categories).Error
	return categories, err
}

// GetDistinctEvents returns all unique events
func (r *ProjectRepository) GetDistinctEvents() ([]string, error) {
	var events []string
	err := r.db.Model(&models.Project{}).Distinct("event").Pluck("event", &events).Error
	return events, err
}

// GetDistinctAwards returns all unique awards
func (r *ProjectRepository) GetDistinctAwards() ([]string, error) {
	var awards []string
	err := r.db.Model(&models.Project{}).Distinct("award").Pluck("award", &awards).Error
	return awards, err
}
