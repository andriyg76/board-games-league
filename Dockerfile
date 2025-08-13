# Stage 1: Build the frontend
FROM --platform=$BUILDPLATFORM node:16 AS frontend-builder

WORKDIR /app/frontend

# Install dependencies and build the frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build the backend
FROM --platform=$BUILDPLATFORM golang:1.24 AS backend-builder

ARG TARGETARCH
ARG TARGETOS

WORKDIR /app

# Install Go dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Run tests
RUN go test ./...

# copy frontend
COPY --from=frontend-builder /app/frontend/dist ./frontendfs

# Build the backend
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /app/main -v .

# Stage 3: Create the final image
FROM ubuntu:latest

WORKDIR /app

# Install necessary packages
RUN apt-get update && apt-get install -y ca-certificates

# Copy the backend binary
COPY --from=backend-builder /app/main .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the backend
CMD ["./main"]