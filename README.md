# Monad Developer Hub Backend

A Go-based REST API for the Monad Developer Hub, built with Gin, GORM, and PostgreSQL.

## Features

- **Project Management**: Submit, review, and manage community projects
- **Submission System**: Unique submission ID tracking with status management
- **Analytics**: Real-time blockchain analytics and transaction data
- **Pagination**: Efficient data pagination for all endpoints
- **Filtering**: Advanced filtering and search capabilities
- **Rate Limiting**: Request rate limiting for API protection
- **CORS**: Cross-origin resource sharing support

## Tech Stack

- **Go 1.21**
- **Gin** - HTTP web framework
- **GORM** - ORM library
- **PostgreSQL** - Database
- **Air** - Live reload for development (optional)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd monad-devhub-be
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. **Create PostgreSQL database**
   ```sql
   CREATE DATABASE monad_devhub;
   ```

5. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```

The API will start on `http://localhost:8080`

## Environment Configuration

Create a `.env` file based on `env.example`:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=monad_devhub

# Server Configuration
PORT=8080
GIN_MODE=debug

# Authentication Configuration
DEFAULT_ADMIN_PASSWORD=admin123
JWT_SECRET=your-super-secret-jwt-key
ADMIN_PASSWORD=admin123  # Legacy fallback

# CORS Configuration
CORS_ORIGINS=http://localhost:3000,http://localhost:3001

# Rate Limiting
RATE_LIMIT_PER_MINUTE=100
```

## API Endpoints

### Health Check
- `GET /health` - Service health check

### Projects
- `GET /api/v1/projects` - Get projects with pagination and filtering
- `GET /api/v1/projects/:id` - Get project by ID
- `POST /api/v1/projects/:id/like` - Like a project

### Submissions ⭐ **Core Feature**
- `POST /api/v1/submissions` - Submit a project (generates submission ID)
- `GET /api/v1/submissions/:submissionId` - Get submission status by ID
- `GET /api/v1/submissions` - Get all submissions
- `PUT /api/v1/submissions/:submissionId/review` - Review submission

### Analytics
- `GET /api/v1/analytics/stats` - Get blockchain statistics
- `GET /api/v1/analytics/transactions` - Get transaction data
- `GET /api/v1/analytics/contracts/top` - Get top contracts

### Authentication 🔐
- `POST /api/v1/auth/login` - Admin login (requires username-password format)
- `GET /api/v1/auth/verify` - Verify JWT token
- `PUT /api/v1/auth/change-password` - Change admin password (protected)
- `POST /api/v1/auth/admin` - Create new admin user (protected)

## Authentication System

The authentication system uses a username-password approach with bcrypt hashing. The frontend uses a single input field with the format `username-password`, and the backend parses it to extract the username and password.

### Default Admin User

On startup, the system automatically creates a default admin user if no admin users exist:

- **Username:** `admin`  
- **Default Password:** `admin123` (or set via `DEFAULT_ADMIN_PASSWORD` environment variable)
- **Frontend Format:** `admin-admin123`

### Input Format

**Frontend Input Format:** `username-password`

**Examples:**
- `admin-admin123` (default)
- `admin-mynewpassword456`
- `john-secretpass789`

### Password Management

**Generate Password Hash:**
```bash
go run scripts/hash_password.go
# Enter password when prompted
```

**Manual User Creation:**
```sql
INSERT INTO admin_users (username, password, is_active, created_at, updated_at)
VALUES ('newuser', '<bcrypt_hash>', true, NOW(), NOW());
```

### Security Features

✅ **Bcrypt Password Hashing** - All passwords securely hashed  
✅ **JWT Authentication** - Secure token-based authentication  
✅ **Protected Routes** - Admin operations require valid JWT tokens  
✅ **Default User Creation** - Automatic setup with secure defaults  
✅ **Password Change** - Dynamic password updates without restart

## Submission ID System

The core feature of this backend is the submission ID system:

### Format
```
SUB-{timestamp}-{randomHash}
Example: SUB-1749035470531-4W6UZJ
```

### Flow
1. **User submits project** → `POST /api/v1/submissions`
2. **Backend generates unique submission ID** → `SUB-1749035470531-4W6UZJ`
3. **Returns submission ID to user** → User can track status
4. **User checks status** → `GET /api/v1/submissions/SUB-1749035470531-4W6UZJ`

### Example Submission Request
```json
POST /api/v1/submissions
{
  "photoLink": "https://example.com/logo.png",
  "projectName": "MonadSwap",
  "description": "Decentralized exchange built on Monad",
  "event": "Mission: 1 Crazy Contract",
  "categories": ["DeFi", "Infrastructure"],
  "teamMembers": [
    {"name": "Alex", "twitter": "alex_dev"},
    {"name": "Sarah", "twitter": "sarah_blockchain"}
  ],
  "playLink": "https://monadswap.example.com",
  "howToPlay": "Connect your wallet and start trading"
}
```

### Example Submission Response
```json
{
  "success": true,
  "submissionId": "SUB-1749035470531-4W6UZJ",
  "message": "Your project has been submitted successfully!",
  "estimatedReviewTime": "2-3 business days",
  "nextSteps": [
    "We'll review your submission within 2-3 business days",
    "You'll receive an email update when review is complete",
    "Use submission ID SUB-1749035470531-4W6UZJ to check status anytime"
  ]
}
```

## Database Schema

The application uses PostgreSQL with GORM for ORM. Key tables:

- `projects` - Approved projects
- `team_members` - Project team members  
- `submissions` - Project submissions (with submission IDs)
- `admin_users` - Admin user credentials (bcrypt hashed passwords)
- `analytics_stats` - Blockchain statistics
- `transactions` - Transaction data
- `contracts` - Smart contract information
- `contract_stats` - Contract statistics

## Error Handling

All APIs return consistent error responses:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": "Additional error details",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Common Error Codes
- `BAD_REQUEST` - Invalid request data
- `DUPLICATE_PROJECT_NAME` - Project name already exists
- `DUPLICATE_SUBMISSION` - Submission already exists
- `INVALID_SUBMISSION_ID` - Invalid submission ID format
- `SUBMISSION_NOT_FOUND` - Submission not found
- `RATE_LIMITED` - Too many requests

## Development

### Project Structure
```
monad-devhub-be/
├── cmd/api/                 # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── database/           # Database connection & migrations
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/            # Data models
│   ├── repository/        # Data access layer
│   ├── services/          # Business logic
│   └── utils/             # Utility functions
├── go.mod                 # Go modules
├── go.sum                 # Dependencies
├── env.example           # Environment template
└── README.md             # This file
```

### Live Reload (Optional)
Install Air for live reload during development:
```bash
go install github.com/cosmtrek/air@latest
air
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License. 