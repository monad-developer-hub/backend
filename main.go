package main

import (
	"log"
	"os"

	"monad-devhub-be/internal/config"
	"monad-devhub-be/internal/database"
	"monad-devhub-be/internal/handlers"
	"monad-devhub-be/internal/middleware"
	"monad-devhub-be/internal/repository"
	"monad-devhub-be/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Initialize services
	projectService := services.NewProjectService(projectRepo, submissionRepo)
	submissionService := services.NewSubmissionService(submissionRepo, projectRepo)
	analyticsService := services.NewAnalyticsService(analyticsRepo)

	// Initialize handlers
	projectHandler := handlers.NewProjectHandler(projectService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	submissionHandler := handlers.NewSubmissionHandler(projectService, submissionService)

	// Setup router
	router := gin.Default()

	// CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORSOrigins
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimit(cfg.RateLimitPerMinute))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "monad-devhub-api"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Projects routes
		projects := v1.Group("/projects")
		{
			projects.GET("", projectHandler.GetProjects)
			projects.POST("", projectHandler.CreateProject)
			projects.GET("/:id", projectHandler.GetProject)
			projects.POST("/:id/like", projectHandler.LikeProject)
		}

		// Submissions routes
		submissions := v1.Group("/submissions")
		{
			submissions.POST("", submissionHandler.SubmitProject)
			submissions.GET("/:submissionId", submissionHandler.GetSubmissionStatus)
			submissions.GET("", submissionHandler.GetSubmissions)
			submissions.PUT("/:submissionId/review", submissionHandler.ReviewSubmission)
		}

		// Analytics routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/stats", analyticsHandler.GetStats)
			analytics.GET("/transactions", analyticsHandler.GetTransactions)
			analytics.GET("/contracts/top", analyticsHandler.GetTopContracts)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("API available at: http://localhost:%s/api/v1", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
