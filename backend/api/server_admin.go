package api

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andriyg76/bgl/utils"
	log "github.com/andriyg76/glog"
	"github.com/go-chi/chi/v5"
)

// Global state for debug logging
var (
	debugLoggingState struct {
		sync.RWMutex
		enabled   bool
		expiresAt *time.Time
		cancelFunc context.CancelFunc
		serverWriter io.Writer
		debugWriter  io.Writer
	}
)

// SetLogWriters stores references to log writers for runtime level changes
func SetLogWriters(serverWriter, debugWriter io.Writer) {
	debugLoggingState.Lock()
	defer debugLoggingState.Unlock()
	debugLoggingState.serverWriter = serverWriter
	debugLoggingState.debugWriter = debugWriter
}

type DebugLoggingState struct {
	Enabled   bool       `json:"enabled"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type EnableDebugRequest struct {
	DurationMinutes int `json:"duration_minutes"`
}

type ServerAdminHandler struct{}

func NewServerAdminHandler() *ServerAdminHandler {
	return &ServerAdminHandler{}
}

func (h *ServerAdminHandler) RegisterRoutes(r chi.Router) {
	r.Post("/debug/enable", h.EnableDebugLogging)
	r.Post("/debug/disable", h.DisableDebugLogging)
	r.Get("/debug/status", h.GetDebugStatus)
	r.Get("/logs/download", h.DownloadLogsByPeriod)
	r.Get("/logs/download-full", h.DownloadFullLogs)
}

func (h *ServerAdminHandler) EnableDebugLogging(w http.ResponseWriter, r *http.Request) {
	var req EnableDebugRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, err, "invalid request body")
		return
	}

	if req.DurationMinutes <= 0 {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, nil, "duration_minutes must be positive")
		return
	}

	debugLoggingState.Lock()
	defer debugLoggingState.Unlock()

	// Cancel existing timer if any
	if debugLoggingState.cancelFunc != nil {
		debugLoggingState.cancelFunc()
	}

	// Set log level to DEBUG
	if debugLoggingState.serverWriter != nil {
		log.SetWriters(debugLoggingState.serverWriter, debugLoggingState.serverWriter, log.DEBUG)
		log.Info("Debug logging enabled for %d minutes", req.DurationMinutes)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(req.DurationMinutes) * time.Minute)
	debugLoggingState.enabled = true
	debugLoggingState.expiresAt = &expiresAt

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.DurationMinutes)*time.Minute)
	debugLoggingState.cancelFunc = cancel

	// Start goroutine to disable after timeout
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			h.disableDebugLoggingInternal()
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DebugLoggingState{
		Enabled:   true,
		ExpiresAt: &expiresAt,
	})
}

func (h *ServerAdminHandler) DisableDebugLogging(w http.ResponseWriter, r *http.Request) {
	h.disableDebugLoggingInternal()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DebugLoggingState{
		Enabled: false,
	})
}

func (h *ServerAdminHandler) disableDebugLoggingInternal() {
	debugLoggingState.Lock()
	defer debugLoggingState.Unlock()

	if debugLoggingState.cancelFunc != nil {
		debugLoggingState.cancelFunc()
		debugLoggingState.cancelFunc = nil
	}

	// Set log level back to INFO
	if debugLoggingState.serverWriter != nil {
		log.SetWriters(debugLoggingState.serverWriter, debugLoggingState.serverWriter, log.INFO)
		log.Info("Debug logging disabled")
	}

	debugLoggingState.enabled = false
	debugLoggingState.expiresAt = nil
}

func (h *ServerAdminHandler) GetDebugStatus(w http.ResponseWriter, r *http.Request) {
	debugLoggingState.RLock()
	defer debugLoggingState.RUnlock()

	state := DebugLoggingState{
		Enabled: debugLoggingState.enabled,
	}
	if debugLoggingState.expiresAt != nil {
		state.ExpiresAt = debugLoggingState.expiresAt
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *ServerAdminHandler) DownloadLogsByPeriod(w http.ResponseWriter, r *http.Request) {
	durationStr := r.URL.Query().Get("duration_minutes")
	filesStr := r.URL.Query().Get("files")

	if durationStr == "" {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, nil, "duration_minutes parameter required")
		return
	}

	var durationMinutes int
	if _, err := fmt.Sscanf(durationStr, "%d", &durationMinutes); err != nil || durationMinutes <= 0 {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, nil, "invalid duration_minutes")
		return
	}

	files := parseLogFiles(filesStr)
	if len(files) == 0 {
		files = []string{"server", "debug"} // default
	}

	logDir := strings.TrimSpace(os.Getenv("LOG_DIR"))
	if logDir == "" {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, nil, "LOG_DIR is not configured")
		return
	}

	startTime := time.Now().Add(-time.Duration(durationMinutes) * time.Minute)
	var result strings.Builder

	for _, fileType := range files {
		logPath := filepath.Join(logDir, fileType+".log")
		lines, err := readLogsSince(logPath, startTime)
		if err != nil {
			// Log error but continue with other files
			log.Warn("Failed to read %s: %v", logPath, err)
			continue
		}

		if len(lines) > 0 {
			result.WriteString(fmt.Sprintf("=== %s.log ===\n", fileType))
			for _, line := range lines {
				result.WriteString(line)
				result.WriteString("\n")
			}
			result.WriteString("\n")
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=logs-%dmin-%s.txt", durationMinutes, time.Now().Format("20060102-150405")))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.String()))
}

func (h *ServerAdminHandler) DownloadFullLogs(w http.ResponseWriter, r *http.Request) {
	filesStr := r.URL.Query().Get("files")
	files := parseLogFiles(filesStr)
	if len(files) == 0 {
		files = []string{"server", "debug"} // default
	}

	logDir := strings.TrimSpace(os.Getenv("LOG_DIR"))
	if logDir == "" {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, nil, "LOG_DIR is not configured")
		return
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, fileType := range files {
		// Add current log file
		currentPath := filepath.Join(logDir, fileType+".log")
		if err := addFileToZip(zipWriter, currentPath, fileType+".log"); err != nil {
			log.Warn("Failed to add %s: %v", currentPath, err)
		}

		// Find and add archived files
		archivedFiles, err := findArchivedLogFiles(logDir, fileType)
		if err != nil {
			log.Warn("Failed to find archived files for %s: %v", fileType, err)
			continue
		}

		for _, archivedPath := range archivedFiles {
			baseName := filepath.Base(archivedPath)
			// Remove .gz extension and add date prefix
			if strings.HasSuffix(baseName, ".gz") {
				baseName = strings.TrimSuffix(baseName, ".gz")
			}
			if err := addGzipFileToZip(zipWriter, archivedPath, baseName); err != nil {
				log.Warn("Failed to add archived file %s: %v", archivedPath, err)
			}
		}
	}

	if err := zipWriter.Close(); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to create zip archive")
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=logs-full-%s.zip", time.Now().Format("20060102-150405")))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func parseLogFiles(filesStr string) []string {
	if filesStr == "" {
		return nil
	}
	parts := strings.Split(filesStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func readLogsSince(logPath string, since time.Time) ([]string, error) {
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() == 0 {
		return []string{}, nil
	}

	// Read entire file (for log files this should be manageable)
	// For very large files, we could optimize by reading from end
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var lines []string
	var currentLine strings.Builder
	collecting := false

	for _, b := range content {
		if b == '\n' {
			if currentLine.Len() > 0 {
				line := currentLine.String()
				currentLine.Reset()

				lineTime := parseLogLineTime(line)
				if !lineTime.IsZero() {
					if lineTime.After(since) || lineTime.Equal(since) {
						collecting = true
						lines = append(lines, line)
					} else if lineTime.Before(since) {
						// Older line - if we're already collecting, this shouldn't happen
						// in a chronological log file, but continue anyway
					}
				} else {
					// If we can't parse timestamp, include line if we're already collecting
					if collecting {
						lines = append(lines, line)
					}
				}
			}
		} else {
			currentLine.WriteByte(b)
		}
	}

	// Handle last line if file doesn't end with newline
	if currentLine.Len() > 0 {
		line := currentLine.String()
		lineTime := parseLogLineTime(line)
		if !lineTime.IsZero() {
			if lineTime.After(since) || lineTime.Equal(since) {
				lines = append(lines, line)
			}
		} else if collecting {
			lines = append(lines, line)
		}
	}

	return lines, nil
}

func parseLogLineTime(line string) time.Time {
	// Try common log formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05.000000",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}

	// Look for timestamp in first 30 characters
	searchLen := len(line)
	if searchLen > 30 {
		searchLen = 30
	}
	prefix := line[:searchLen]

	for _, format := range formats {
		// Try to find timestamp at different positions
		for i := 0; i <= len(prefix)-len(format); i++ {
			if t, err := time.Parse(format, prefix[i:i+len(format)]); err == nil {
				return t
			}
		}
	}

	return time.Time{}
}

func addFileToZip(zipWriter *zip.Writer, filePath, zipName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, skip
		}
		return err
	}
	defer file.Close()

	zipEntry, err := zipWriter.Create(zipName)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipEntry, file)
	return err
}

func addGzipFileToZip(zipWriter *zip.Writer, gzipPath, zipName string) error {
	gzipFile, err := os.Open(gzipPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer gzipFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	zipEntry, err := zipWriter.Create(zipName)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipEntry, gzipReader)
	return err
}

func findArchivedLogFiles(logDir, fileType string) ([]string, error) {
	pattern := filepath.Join(logDir, fileType+"-*.log.gz")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Also check for non-gzipped archived files
	pattern2 := filepath.Join(logDir, fileType+"-*.log")
	matches2, err := filepath.Glob(pattern2)
	if err == nil {
		matches = append(matches, matches2...)
	}

	return matches, nil
}
