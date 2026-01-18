# Envie Backend

Go API server for Envie.

## Tech Stack

- **[Go](https://go.dev/)** - Programming language
- **[Gin](https://gin-gonic.com/)** - HTTP web framework
- **[GORM](https://gorm.io/)** - ORM for PostgreSQL
- **[AWS SDK v2](https://aws.github.io/aws-sdk-go-v2/)** - S3-compatible storage
- **[golang-jwt](https://github.com/golang-jwt/jwt)** - JWT authentication

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go      # Entry point
├── internal/
│   ├── auth/
│   │   ├── jwt.go       # JWT token generation/validation
│   │   └── oauth.go     # GitHub OAuth
│   ├── database/
│   │   └── db.go        # PostgreSQL connection
│   ├── handlers/
│   │   ├── auth.go          # Auth endpoints
│   │   ├── config.go        # Config items (env vars)
│   │   ├── file.go          # File upload/download
│   │   ├── identity.go      # Device management
│   │   ├── key_rotation.go  # Key rotation flow
│   │   ├── organization.go  # Organization CRUD
│   │   ├── project.go       # Project CRUD
│   │   ├── secret_manager.go # GCP Secret Manager integration
│   │   ├── team.go          # Team management
│   │   └── user.go          # User endpoints
│   ├── middleware/
│   │   └── auth.go      # JWT auth middleware
│   ├── models/
│   │   ├── config.go        # ConfigItem model
│   │   ├── file.go          # ProjectFile model
│   │   ├── identity.go      # UserIdentity model
│   │   ├── key_rotation.go  # KeyRotation models
│   │   ├── linking_code.go  # LinkingCode, RefreshToken
│   │   ├── organization.go  # Organization, OrganizationUser
│   │   ├── project.go       # Project, TeamProject
│   │   ├── secret_manager.go # SecretManagerConfig
│   │   ├── team.go          # Team, TeamUser
│   │   └── user.go          # User model
│   └── storage/
│       └── s3.go        # S3-compatible file storage
├── go.mod
├── go.sum
└── Dockerfile
```

## API Endpoints

### Public
- `GET /auth/login` - Initiate GitHub OAuth
- `GET /auth/callback` - OAuth callback
- `POST /auth/exchange` - Exchange linking code for tokens
- `POST /auth/refresh` - Refresh access token

### Protected (require Bearer token)

**User**
- `GET /me` - Get current user
- `POST /auth/logout` - Logout

**Devices/Identity**
- `GET /devices` - List user's devices
- `POST /devices` - Register new device
- `PUT /devices/:id` - Update device (approve with encrypted master key)
- `DELETE /devices/:id` - Delete device

**Projects**
- `GET /projects` - List projects
- `POST /projects` - Create project
- `GET /projects/:id` - Get project
- `PUT /projects/:id` - Update project
- `DELETE /projects/:id` - Delete project
- `GET /projects/:id/config` - Get config items
- `PUT /projects/:id/config` - Sync config items

**Files**
- `GET /projects/:id/files` - List files
- `POST /projects/:id/files` - Upload file
- `GET /projects/:id/files/:fileId` - Download file
- `DELETE /projects/:id/files/:fileId` - Delete file

**Teams & Organizations**
- `GET /organizations` - List organizations
- `POST /organizations` - Create organization
- `GET /teams` - List teams
- `POST /teams` - Create team
- `GET /teams/:id/members` - List team members
- `POST /teams/:id/members` - Add member
- `DELETE /teams/:id/members/:userId` - Remove member

**Key Rotation**
- `POST /projects/:id/rotation` - Initiate rotation
- `POST /projects/:id/rotation/:rotationId/approve` - Approve rotation
- `POST /projects/:id/rotation/:rotationId/reject` - Reject rotation

## Environment Variables

Create a `.env` file in the backend directory:

```bash
# Database (required)
DB_DSN=postgres://user:password@localhost:5432/envie?sslmode=disable

# JWT (required)
JWT_SECRET=your-secret-key-min-32-chars

# GitHub OAuth (required)
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_REDIRECT_URL=http://localhost:8080/auth/callback

# S3 Storage - Tigris (required)
TIGRIS_STORAGE_ACCESS_KEY_ID=your-access-key
TIGRIS_STORAGE_SECRET_ACCESS_KEY=your-secret-key
TIGRIS_STORAGE_ENDPOINT=https://fly.storage.tigris.dev
TIGRIS_BUCKET_NAME=your-bucket-name
```

### Variable Details

| Variable | Description |
|----------|-------------|
| `DB_DSN` | PostgreSQL connection string |
| `JWT_SECRET` | Secret for signing JWT tokens (min 32 characters recommended) |
| `GITHUB_CLIENT_ID` | GitHub OAuth App client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth App client secret |
| `GITHUB_REDIRECT_URL` | OAuth callback URL |
| `TIGRIS_STORAGE_ACCESS_KEY_ID` | S3 access key (Tigris, AWS, etc.) |
| `TIGRIS_STORAGE_SECRET_ACCESS_KEY` | S3 secret key |
| `TIGRIS_STORAGE_ENDPOINT` | S3 endpoint URL |
| `TIGRIS_BUCKET_NAME` | S3 bucket name for file storage |

## Development

```bash
# Install dependencies
go mod download

# Run server
go run cmd/api/main.go

# Build binary
go build -o envie-backend cmd/api/main.go
```

The server runs on port `8080` by default.

## Database

Uses PostgreSQL with GORM. Migrations run automatically on startup.

## Docker

```bash
docker build -t envie-backend .
docker run -p 8080:8080 --env-file .env envie-backend
```
