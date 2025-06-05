package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Project represents an approved project in the database
type Project struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"uniqueIndex;not null"`
	Logo         string         `json:"logo"`
	Description  string         `json:"description" gorm:"not null"`
	Categories   pq.StringArray `json:"categories" gorm:"type:text[]"`
	Event        string         `json:"event" gorm:"not null"`
	Award        string         `json:"award"`
	Likes        int            `json:"likes" gorm:"default:0"`
	Comments     int            `json:"comments" gorm:"default:0"`
	HowToPlay    string         `json:"howToPlay" gorm:"column:how_to_play;not null"`
	PlayURL      string         `json:"playUrl" gorm:"column:play_url;not null"`
	GithubURL    *string        `json:"github,omitempty" gorm:"column:github_url"`
	WebsiteURL   *string        `json:"website,omitempty" gorm:"column:website_url"`
	TeamMembers  []TeamMember   `json:"team" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	SubmissionID *string        `json:"submissionId,omitempty" gorm:"column:submission_id;uniqueIndex"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TeamMember represents a project team member
type TeamMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"projectId" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null"`
	Twitter   string    `json:"twitter"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"createdAt"`
}

// Submission represents a project submission awaiting review
type Submission struct {
	ID                string         `json:"id" gorm:"primaryKey"` // Will be the submission ID like SUB-xxx
	ProjectName       string         `json:"projectName" gorm:"column:project_name;not null"`
	Description       string         `json:"description" gorm:"not null"`
	PhotoLink         string         `json:"photoLink" gorm:"column:photo_link"`
	Event             string         `json:"event" gorm:"not null"`
	Categories        pq.StringArray `json:"categories" gorm:"type:text[]"`
	TeamMembers       string         `json:"teamMembers" gorm:"column:team_members;type:jsonb"` // Store as JSON
	GithubLink        *string        `json:"githubLink,omitempty" gorm:"column:github_link"`
	WebsiteLink       *string        `json:"websiteLink,omitempty" gorm:"column:website_link"`
	PlayLink          string         `json:"playLink" gorm:"column:play_link;not null"`
	HowToPlay         string         `json:"howToPlay" gorm:"column:how_to_play;not null"`
	AdditionalNotes   *string        `json:"additionalNotes,omitempty" gorm:"column:additional_notes"`
	Status            string         `json:"status" gorm:"default:'pending'"`
	ReviewerID        *uint          `json:"reviewerId,omitempty" gorm:"column:reviewer_id"`
	Feedback          *string        `json:"feedback,omitempty"`
	ChangesRequested  pq.StringArray `json:"changesRequested,omitempty" gorm:"column:changes_requested;type:text[]"`
	SubmittedAt       time.Time      `json:"submittedAt" gorm:"column:submitted_at"`
	ReviewStartedAt   *time.Time     `json:"reviewStartedAt,omitempty" gorm:"column:review_started_at"`
	ReviewedAt        *time.Time     `json:"reviewedAt,omitempty" gorm:"column:reviewed_at"`
	PublishedAt       *time.Time     `json:"publishedAt,omitempty" gorm:"column:published_at"`
	ApprovedProjectID *uint          `json:"approvedProjectId,omitempty" gorm:"column:approved_project_id"`
	ApprovedProject   *Project       `json:"project,omitempty" gorm:"foreignKey:ApprovedProjectID"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

// AnalyticsStats represents blockchain analytics statistics
type AnalyticsStats struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	TotalTransactions int64     `json:"totalTransactions" gorm:"column:total_transactions"`
	TPS               int       `json:"tps"`
	ActiveValidators  int       `json:"activeValidators" gorm:"column:active_validators"`
	BlockHeight       int64     `json:"blockHeight" gorm:"column:block_height"`
	Timestamp         time.Time `json:"timestamp"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Hash        string    `json:"hash" gorm:"uniqueIndex;not null"`
	Type        string    `json:"type" gorm:"not null"`
	FromAddress string    `json:"from" gorm:"column:from_address;not null"`
	ToAddress   string    `json:"to" gorm:"column:to_address;not null"`
	Value       float64   `json:"value"`
	GasUsed     int64     `json:"gasUsed" gorm:"column:gas_used"`
	BlockNumber int64     `json:"blockNumber" gorm:"column:block_number"`
	Timestamp   time.Time `json:"timestamp"`
}

// Contract represents a smart contract on the blockchain
type Contract struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Logo      string    `json:"logo"`
	Address   string    `json:"address" gorm:"uniqueIndex;not null"`
	Category  string    `json:"category"`
	Verified  bool      `json:"verified" gorm:"default:false"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ContractStats represents aggregated contract statistics
type ContractStats struct {
	ContractID    string    `json:"contractId" gorm:"column:contract_id;primaryKey"`
	Contract      Contract  `json:"contract" gorm:"foreignKey:ContractID"`
	TxCount       int       `json:"txCount" gorm:"column:tx_count"`
	UniqueWallets int       `json:"uniqueWallets" gorm:"column:unique_wallets"`
	Change24h     float64   `json:"change24h" gorm:"column:change_24h"`
	GasUsed       int64     `json:"gasUsed" gorm:"column:gas_used"`
	LastUpdated   time.Time `json:"lastUpdated" gorm:"column:last_updated"`
}

// TeamMemberInput represents team member data from submission form
type TeamMemberInput struct {
	Name    string `json:"name" binding:"required"`
	Twitter string `json:"twitter" binding:"required"`
}

// AdminUser represents admin users with username/password authentication
type AdminUser struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Password  string    `json:"password" gorm:"not null"` // In production, this should be hashed
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
