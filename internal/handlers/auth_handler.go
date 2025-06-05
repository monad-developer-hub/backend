package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"monad-devhub-be/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	handler := &AuthHandler{db: db}
	// Initialize default admin user
	handler.initializeDefaultAdmin()
	return handler
}

type LoginRequest struct {
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

type CreateAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	Role     string `json:"role"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// initializeDefaultAdmin creates a default admin user if none exists
func (h *AuthHandler) initializeDefaultAdmin() {
	var count int64
	h.db.Model(&models.AdminUser{}).Count(&count)

	if count == 0 {
		// Create default admin user
		defaultPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
		if defaultPassword == "" {
			defaultPassword = "admin123" // Default password
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return // Silently fail for now
		}

		defaultAdmin := models.AdminUser{
			Username: "admin",
			Password: string(hashedPassword),
			IsActive: true,
		}

		h.db.Create(&defaultAdmin)
	}
}

// parseCredentials parses the format "username-password" from frontend input
func (h *AuthHandler) parseCredentials(input string) (username, password string, valid bool) {
	// Find the first dash separator
	dashIndex := strings.Index(input, "-")
	if dashIndex == -1 || dashIndex == 0 || dashIndex == len(input)-1 {
		return "", "", false
	}

	username = input[:dashIndex]
	password = input[dashIndex+1:]

	// Basic validation
	if len(username) < 2 || len(password) < 3 {
		return "", "", false
	}

	return username, password, true
}

// authenticateUser validates username/password against database with fallback to env
func (h *AuthHandler) authenticateUser(username, password string) (bool, string) {
	var adminUser models.AdminUser
	err := h.db.Where("username = ? AND is_active = ?", username, true).First(&adminUser).Error

	if err == nil {
		// User found in database - check hashed password
		err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(password))
		if err == nil {
			return true, username
		}
		return false, ""
	}

	// Fallback to environment variable for backward compatibility
	envPassword := os.Getenv("ADMIN_PASSWORD")
	if envPassword == "" {
		envPassword = "admin123" // Default for development
	}

	// Check if it matches the legacy format (admin-password)
	if username == "admin" && password == envPassword {
		return true, "admin"
	}

	return false, ""
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	// Parse username-password format
	username, password, valid := h.parseCredentials(req.Password)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_FORMAT",
				"message": "Invalid credential format. Expected: username-password",
			},
		})
		return
	}

	// Authenticate user
	authenticated, authenticatedUsername := h.authenticateUser(username, password)
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_CREDENTIALS",
				"message": "Invalid username or password",
			},
		})
		return
	}

	// Generate JWT token
	token, err := h.generateJWT(authenticatedUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "TOKEN_GENERATION_FAILED",
				"message": "Failed to generate authentication token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   token,
		Message: "Login successful",
	})
}

// CreateAdmin handles POST /api/v1/auth/admin (protected endpoint)
func (h *AuthHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	// Check if username already exists
	var existingUser models.AdminUser
	err := h.db.Where("username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "USERNAME_EXISTS",
				"message": "Username already exists",
			},
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "HASH_FAILED",
				"message": "Failed to hash password",
			},
		})
		return
	}

	// Create new admin user
	adminUser := models.AdminUser{
		Username: req.Username,
		Password: string(hashedPassword),
		IsActive: true,
	}

	if err := h.db.Create(&adminUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CREATE_FAILED",
				"message": "Failed to create admin user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Admin user created successfully",
		"username": adminUser.Username,
	})
}

// ChangePassword handles PUT /api/v1/auth/change-password (protected endpoint)
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	// Parse current credentials
	currentUsername, currentPassword, valid := h.parseCredentials(req.CurrentPassword)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_FORMAT",
				"message": "Invalid current credential format. Expected: username-password",
			},
		})
		return
	}

	// Parse new credentials
	newUsername, newPassword, valid := h.parseCredentials(req.NewPassword)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_FORMAT",
				"message": "Invalid new credential format. Expected: username-password",
			},
		})
		return
	}

	// Verify current credentials
	authenticated, _ := h.authenticateUser(currentUsername, currentPassword)
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_CREDENTIALS",
				"message": "Invalid current credentials",
			},
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "HASH_FAILED",
				"message": "Failed to hash new password",
			},
		})
		return
	}

	// Update or create admin user
	var adminUser models.AdminUser
	err = h.db.Where("username = ?", newUsername).First(&adminUser).Error

	if err != nil {
		// Create new admin user
		adminUser = models.AdminUser{
			Username: newUsername,
			Password: string(hashedPassword),
			IsActive: true,
		}
		if err := h.db.Create(&adminUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UPDATE_FAILED",
					"message": "Failed to create new admin user",
				},
			})
			return
		}
	} else {
		// Update existing admin user
		adminUser.Password = string(hashedPassword)
		if err := h.db.Save(&adminUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UPDATE_FAILED",
					"message": "Failed to update admin user",
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Credentials updated successfully",
	})
}

// VerifyToken handles GET /api/v1/auth/verify
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NO_TOKEN",
				"message": "No authorization token provided",
			},
		})
		return
	}

	// Remove "Bearer " prefix
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Validate token
	if h.validateJWT(tokenString) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Token is valid",
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_TOKEN",
				"message": "Token is invalid or expired",
			},
		})
	}
}

// generateJWT creates a new JWT token with 24h expiration
func (h *AuthHandler) generateJWT(username string) (string, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key" // Default for development
	}

	// Create claims
	claims := Claims{
		Role:     "admin",
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "monad-devhub-api",
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// validateJWT validates a JWT token
func (h *AuthHandler) validateJWT(tokenString string) bool {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key"
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return false
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is expired
		return claims.ExpiresAt.After(time.Now())
	}

	return false
}
