# Release Process

*[Українська версія](RELEASE.md)*

## Overview

The release process is automated through a GitHub Actions workflow. It allows creating a new release with automatic version bumping, Docker image building, tag creation, GitHub Release, and automatic deployment to the production server.

## Running a Release

### Via GitHub UI

1. Navigate to Actions page: https://github.com/andriyg76/board-games-league/actions/workflows/release.yml
2. Click "Run workflow"
3. Select the branch (`main` or `release/v*`)
4. Select the version bump type:
   - **major** - increments major version (1.0.0 → 2.0.0)
   - **minor** - increments minor version (1.0.0 → 1.1.0)
   - **patch** - increments patch version (1.0.0 → 1.0.1)
5. Click "Run workflow"

### Branch Requirements

The workflow can only be run from:
- The `main` branch
- Branches matching `release/v*` pattern (e.g., `release/v1.0`)

If you try to run it from another branch, the workflow will fail.

## What Happens During Release

1. **Branch validation** - checks that the workflow is run from the correct branch
2. **Get latest version** - finds the latest tag matching `v*.*.*` format in the current branch history
3. **Calculate new version** - applies the selected bump type to the latest version
4. **Docker build/push** - builds and pushes Docker image with the new version tag (e.g., `v1.2.3`)
5. **Create tag** - creates a git tag with the new version and pushes it to the repository
6. **Generate changelog** - generates a list of changes between the previous tag and current HEAD
7. **Create GitHub Release** - creates a GitHub Release with the changelog
8. **Deploy to production** - automatically:
   - Updates `BACKEND_VERSION` in `.env` file on the production server
   - Runs `docker compose pull` to download the new image
   - Runs `docker compose up -d` to restart containers
9. **Health check** - verifies that the backend container is running correctly

## Environment Setup

### Required GitHub Secrets

The following secrets need to be configured in GitHub for the workflow to work:

#### Docker Hub
- `DOCKER_USERNAME` - Docker Hub username
- `DOCKER_PASSWORD` - Docker Hub password or access token

#### Production Server (SSH)
- `SSH_HOST` - hostname or IP address of the production server
- `SSH_USER` - SSH username for connection
- `SSH_PRIVATE_KEY` - private SSH key for server access

### How to Configure Secrets

1. Go to repository Settings
2. Select "Secrets and variables" → "Actions"
3. Click "New repository secret"
4. Add each secret with name and value

### Production Server Preparation

The production server must have:

1. **Docker and Docker Compose** installed and configured
2. **Directory `~/board-games-league`** with files:
   - `.env` - file with `BACKEND_VERSION` (will be automatically updated)
   - `.backend.env` - backend configuration (created manually or via other workflows)
   - `.mongo.env` - MongoDB configuration (created manually or via other workflows)
   - `docker-compose.yaml` - docker-compose file with backend and mongo services

3. **SSH access** configured from GitHub Actions runner

Example `docker-compose.yaml`:
```yaml
services:
  backend:
    image: {docker-registry}/{image-name}:${BACKEND_VERSION}
    ports:
      - "20032:8080"
    env_file:
      - .backend.env
    depends_on:
      - mongo

  mongo:
    image: mongodb/mongodb-community-server:8.2-ubuntu2204
    env_file:
      - .mongo.env
    volumes:
      - mongo_data:/data/db
      - mongo_configdb:/data/configdb

volumes:
  mongo_data: {}
  mongo_configdb: {}
```

## Versioning

### Version Format

Versions use [Semantic Versioning](https://semver.org/) format:
- `vMAJOR.MINOR.PATCH` (e.g., `v1.2.3`)

### Latest Version Detection

The workflow finds the latest tag matching `v*.*.*` format that exists in the current branch history. This means if you're working on branch `release/v1.0`, the workflow will only look for tags in that branch's history, not all repository tags.

### BACKEND_VERSION

After release, `BACKEND_VERSION` in the `.env` file on the production server is updated to the full version with `v` prefix:
```
BACKEND_VERSION=v1.2.3
```

## Health Check After Deployment

After deployment, the workflow automatically checks the backend container status:

- Container must be in `running` state
- Container must not be in states: `exited`, `dead`, or `restarting`

If the container is not working correctly, the workflow will fail, preventing the production environment from being left in a broken state.

## Troubleshooting

### Workflow Won't Start

- Check that you're running from `main` or `release/v*` branch
- Check that all required secrets are configured

### Docker Build/Push Fails

- Verify that `DOCKER_USERNAME` and `DOCKER_PASSWORD` are set correctly
- Verify that the user has push permissions to the Docker registry

### Deployment Fails

- Check that SSH secrets are configured correctly
- Verify that the SSH key has access to the production server
- Check that Docker and Docker Compose are installed on the server
- Verify that the `~/board-games-league` directory exists and has correct permissions

### Container Won't Start After Deployment

- Check logs: `docker compose logs backend` on the production server
- Verify that the image with the correct version exists in the Docker registry
- Check that `BACKEND_VERSION` in `.env` is updated correctly
- Verify configuration in `.backend.env` and `.mongo.env`

### Tag Creation Fails

- Check that the workflow has `contents: write` permission
- Verify that the branch has permissions to push tags

## Links

- [GitHub Actions Workflow](https://github.com/andriyg76/board-games-league/actions/workflows/release.yml)
- [Semantic Versioning](https://semver.org/)

