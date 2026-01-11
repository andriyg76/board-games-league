# Stage 1: Build the frontend (on build platform to avoid emulation)
FROM --platform=$BUILDPLATFORM node:22 AS frontend-builder

ARG BUILD_VERSION=unknown
ARG BUILD_COMMIT=unknown
ARG BUILD_BRANCH=unknown
ARG BUILD_DATE

WORKDIR /app/frontend

# Install dependencies and build the frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./

# Create version.json file with build info
RUN echo "{\"version\":\"${BUILD_VERSION}\",\"commit\":\"${BUILD_COMMIT}\",\"branch\":\"${BUILD_BRANCH}\",\"date\":\"${BUILD_DATE}\"}" > public/version.json || echo "{}" > public/version.json

RUN npm run build

# Stage 2: Build the backend for both architectures
FROM --platform=$BUILDPLATFORM golang:1.24 AS backend-builder

ARG BUILD_VERSION=unknown
ARG BUILD_COMMIT=unknown
ARG BUILD_BRANCH=unknown
ARG BUILD_DATE

WORKDIR /app

# Install git for getting commit info if needed
RUN apt-get update && apt-get install -y git || true

# Install Go dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Run tests
RUN go test ./...

# copy frontend
COPY --from=frontend-builder /app/frontend/dist ./frontendfs

# Build the backend for both architectures
RUN BUILD_DATE_VAL="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}" && \
  GOOS=linux GOARCH=amd64 go build \
  -ldflags "-X github.com/andriyg76/bgl/api.BuildVersion=${BUILD_VERSION} -X github.com/andriyg76/bgl/api.BuildCommit=${BUILD_COMMIT} -X github.com/andriyg76/bgl/api.BuildBranch=${BUILD_BRANCH} -X github.com/andriyg76/bgl/api.BuildDate=${BUILD_DATE_VAL}" \
  -o /app/main-amd64 -v . && \
  GOOS=linux GOARCH=arm64 go build \
  -ldflags "-X github.com/andriyg76/bgl/api.BuildVersion=${BUILD_VERSION} -X github.com/andriyg76/bgl/api.BuildCommit=${BUILD_COMMIT} -X github.com/andriyg76/bgl/api.BuildBranch=${BUILD_BRANCH} -X github.com/andriyg76/bgl/api.BuildDate=${BUILD_DATE_VAL}" \
  -o /app/main-arm64 -v .

# Stage 3: Create the final image
FROM ubuntu:latest

ARG TARGETARCH

WORKDIR /app

# Install necessary packages
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy both binaries
COPY --from=backend-builder /app/main-amd64 ./main-amd64
COPY --from=backend-builder /app/main-arm64 ./main-arm64

# Select the correct binary based on target architecture
RUN cp ./main-${TARGETARCH} ./main && rm ./main-amd64 ./main-arm64

# Expose the port the app runs on
EXPOSE 8080

# Command to run the backend
CMD ["./main"]
