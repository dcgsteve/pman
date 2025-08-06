# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PMAN is a secure password manager built in Go with a client-server architecture, designed for team password sharing with enterprise-grade security and automation support.

## Build Commands

```bash
# Build all platform binaries (Linux, macOS, Windows for AMD64/ARM64)
./build-binaries.sh

# Build Docker server image (local only)
./build-docker-server.sh

# Build and push Docker image to Docker Hub
./build-docker-server.sh /p

# Create GitHub releases (needs `gh` CLI tool installed and authenticated)
./create-github-releases.sh
```

## Development Commands

```bash
# Run the server locally (requires environment variables)
go run backend/main.go

# Run the CLI client
go run cli/main.go [command]

# Format Go code
go fmt ./...

# Check for issues
go vet ./...
```

## Testing

Currently, there are no automated tests in this codebase. When adding new features, consider implementing tests using Go's standard testing framework.

## Architecture

### Component Structure
- **`cli/`**: Command-line client that communicates with the backend API
- **`backend/`**: RESTful API server with JWT authentication
- **`shared/`**: Common models, utilities, and configuration shared between client and server

### Key Architectural Decisions

1. **Database**: SQLite with modernc.org/sqlite (pure Go, no CGO required)
2. **Authentication**: JWT tokens with blacklisting capability for immediate revocation
3. **Encryption**: 
   - Server-side: AES-256 encryption for password storage
   - Client-side: Machine-specific encryption for local configuration
   - Transport: HTTPS required in production
4. **API Router**: Gorilla Mux for HTTP routing
5. **Security Model**: Role-based access control with admin/user roles and group-based password sharing

### Database Schema

The schema is defined in `backend/database/schema.sql` with these core tables:
- `users`: User accounts with roles and groups
- `passwords`: Encrypted password entries with audit fields
- `groups`: Group metadata for password sharing
- `tokens`: JWT token tracking and blacklisting

### Environment Configuration

Required environment variables for the server:
- `PMAN_ENCRYPTION_KEY`: Master encryption key for password storage
- `PMAN_DOMAIN_NAME`: Domain for JWT token validation

Optional:
- `PMAN_DEFAULT_EXPIRE_DAYS`: Token expiration (default: 24)
- `PORT`: Server port (default: 5000)

### API Endpoints

The backend provides RESTful endpoints organized by function:
- **Auth**: `/api/auth/*` - Login, logout, token management
- **Passwords**: `/api/passwords/*` - CRUD operations on passwords
- **Users**: `/api/users/*` - User management (admin only)
- **Groups**: `/api/groups/*` - Group management

All endpoints except `/api/auth/login` and `/health` require JWT authentication via Bearer token.

### CLI Command Structure

Commands follow the pattern: `pman [command] [arguments]`

Core command categories:
- Authentication: `login`, `logout`, `passwd`, `whoami`
- Password operations: `add`, `get`, `edit`, `rm`, `ls`, `info`
- Group management: `setgroup`
- Admin functions: `useradd`, `userdel`, `userlist`, etc.

The CLI supports both interactive prompts and piping for automation.

## Code Conventions

- Use Go standard formatting (`go fmt`)
- Follow existing error handling patterns with proper error wrapping
- Maintain separation between CLI, backend, and shared code
- Keep security-sensitive operations in dedicated packages
- Use structured logging with clear context

## Security Considerations

- Never log passwords or encryption keys
- Always validate JWT tokens before processing requests
- Use parameterized queries to prevent SQL injection
- Sanitize all user inputs
- Maintain audit trails for password access