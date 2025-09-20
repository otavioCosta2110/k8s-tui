package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Level int

const (
	LEVEL_DEBUG Level = iota
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
)

const (
	LogDir         = "./local/state/k8s-tui"
	MaxLogFileSize = 10 * 1024 * 1024 // 10MB
	MaxLogFiles    = 5                // Keep 5 rotated files
)

func (l Level) String() string {
	switch l {
	case LEVEL_DEBUG:
		return "DEBUG"
	case LEVEL_INFO:
		return "INFO"
	case LEVEL_WARN:
		return "WARN"
	case LEVEL_ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level  Level
	logger *log.Logger
	file   *os.File
	logDir string
}

var defaultLogger *Logger

func init() {
	// Create log directory
	if err := os.MkdirAll(LogDir, 0755); err != nil {
		log.Printf("Failed to create log directory %s: %v", LogDir, err)
		return
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(LogDir, fmt.Sprintf("k8s-tui-%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	defaultLogger = &Logger{
		level:  LEVEL_DEBUG,
		logger: log.New(file, "", 0), // No prefix, we'll add our own
		file:   file,
		logDir: LogDir,
	}
}

func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.level = level
	}
}

func rotateLogFile() error {
	if defaultLogger == nil || defaultLogger.file == nil {
		return nil
	}

	// Check file size
	stat, err := defaultLogger.file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() < MaxLogFileSize {
		return nil
	}

	// Close current file
	defaultLogger.file.Close()

	// Rotate existing files
	for i := MaxLogFiles - 1; i >= 1; i-- {
		oldFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log.%d", time.Now().Format("2006-01-02"), i))
		newFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log.%d", time.Now().Format("2006-01-02"), i+1))

		if _, err := os.Stat(oldFile); err == nil {
			os.Rename(oldFile, newFile)
		}
	}

	// Move current file to .1
	currentFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log", time.Now().Format("2006-01-02")))
	rotatedFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log.1", time.Now().Format("2006-01-02")))
	os.Rename(currentFile, rotatedFile)

	// Create new log file
	file, err := os.OpenFile(currentFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defaultLogger.file = file
	defaultLogger.logger = log.New(file, "", 0)

	return nil
}

func logMessage(level Level, message string) {
	if defaultLogger == nil || level < defaultLogger.level {
		return
	}

	// Check if rotation is needed
	if err := rotateLogFile(); err != nil {
		// Log rotation error to stderr as fallback
		log.Printf("Failed to rotate log file: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), message)

	defaultLogger.logger.Print(formattedMessage)
}

func Debug(message string) {
	logMessage(LEVEL_DEBUG, message)
}

func Info(message string) {
	logMessage(LEVEL_INFO, message)
}

func Warn(message string) {
	logMessage(LEVEL_WARN, message)
}

func Error(message string) {
	logMessage(LEVEL_ERROR, message)
}

func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}
