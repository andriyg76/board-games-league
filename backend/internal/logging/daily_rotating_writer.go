package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andriyg76/hexerr"
)

const dateFormat = "2006-01-02"

type DailyRotatingWriter struct {
	basePath    string
	currentDate string
	file        *os.File
	mu          sync.Mutex
}

func NewDailyRotatingWriter(basePath string) (*DailyRotatingWriter, error) {
	writer := &DailyRotatingWriter{basePath: basePath}
	if err := writer.init(); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *DailyRotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateIfNeeded(time.Now()); err != nil {
		return 0, err
	}
	if w.file == nil {
		return 0, hexerr.New("log file is not available")
	}
	return w.file.Write(p)
}

func (w *DailyRotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *DailyRotatingWriter) init() error {
	now := time.Now()
	w.currentDate = now.Format(dateFormat)

	if info, err := os.Stat(w.basePath); err == nil {
		w.currentDate = info.ModTime().Format(dateFormat)
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := w.openFile(); err != nil {
		return err
	}

	return w.rotateIfNeeded(now)
}

func (w *DailyRotatingWriter) openFile() error {
	file, err := os.OpenFile(w.basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = file
	return nil
}

func (w *DailyRotatingWriter) rotateIfNeeded(now time.Time) error {
	date := now.Format(dateFormat)
	if w.file == nil {
		w.currentDate = date
		return w.openFile()
	}
	if date == w.currentDate {
		return nil
	}

	oldDate := w.currentDate
	if err := w.rotate(oldDate); err != nil {
		return err
	}
	w.currentDate = date

	return w.openFile()
}

func (w *DailyRotatingWriter) rotate(date string) error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}

	if _, err := os.Stat(w.basePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	archivedPath, err := uniqueArchivePath(w.basePath, date)
	if err != nil {
		return err
	}
	if err := os.Rename(w.basePath, archivedPath); err != nil {
		return err
	}

	return gzipAndRemove(archivedPath)
}

func uniqueArchivePath(basePath, date string) (string, error) {
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	for i := 0; ; i++ {
		suffix := ""
		if i > 0 {
			suffix = fmt.Sprintf("-%d", i)
		}
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%s%s%s", name, date, suffix, ext))
		exists, err := archiveExists(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

func archiveExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if _, err := os.Stat(path + ".gz"); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}

func gzipAndRemove(path string) error {
	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.OpenFile(path+".gz", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(target)

	if _, err := io.Copy(gzipWriter, source); err != nil {
		_ = gzipWriter.Close()
		_ = target.Close()
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		_ = target.Close()
		return err
	}
	if err := target.Close(); err != nil {
		return err
	}

	return os.Remove(path)
}
