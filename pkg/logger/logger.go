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
	MaxLogFileSize = 10 * 1024 * 1024
	MaxLogFiles    = 5
)

func getLogDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory can't be determined
		return "./.local/state/k8s-tui/logs"
	}
	return filepath.Join(homeDir, ".local", "state", "k8s-tui", "logs")
}

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
var pluginLoggers = make(map[string]*Logger)

func init() {
	logDir := getLogDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory %s: %v", logDir, err)
		return
	}

	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("k8s-tui-%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	defaultLogger = &Logger{
		level:  LEVEL_DEBUG,
		logger: log.New(file, "", 0),
		file:   file,
		logDir: logDir,
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

	stat, err := defaultLogger.file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() < MaxLogFileSize {
		return nil
	}

	defaultLogger.file.Close()

	for i := MaxLogFiles - 1; i >= 1; i-- {
		oldFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log.%d", time.Now().Format("2006-01-02"), i))
		newFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log.%d", time.Now().Format("2006-01-02"), i+1))

		if _, err := os.Stat(oldFile); err == nil {
			os.Rename(oldFile, newFile)
		}
	}

	currentFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log", time.Now().Format("2006-01-02")))
	rotatedFile := filepath.Join(defaultLogger.logDir, fmt.Sprintf("k8s-tui-%s.log.1", time.Now().Format("2006-01-02")))
	os.Rename(currentFile, rotatedFile)

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

	if err := rotateLogFile(); err != nil {
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

func GetPluginLogger(pluginName string) *Logger {
	if logger, exists := pluginLoggers[pluginName]; exists {
		return logger
	}

	// Create plugin-specific log directory
	pluginLogDir := filepath.Join(getLogDir(), "plugins")
	if err := os.MkdirAll(pluginLogDir, 0755); err != nil {
		log.Printf("Failed to create plugin log directory %s: %v", pluginLogDir, err)
		return defaultLogger
	}

	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(pluginLogDir, fmt.Sprintf("%s-%s.log", pluginName, timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open plugin log file %s: %v", logFile, err)
		return defaultLogger
	}

	pluginLogger := &Logger{
		level:  LEVEL_DEBUG,
		logger: log.New(file, "", 0),
		file:   file,
		logDir: pluginLogDir,
	}

	pluginLoggers[pluginName] = pluginLogger
	return pluginLogger
}

func PluginDebug(pluginName, message string) {
	logger := GetPluginLogger(pluginName)
	if logger != nil {
		logMessageWithLogger(logger, LEVEL_DEBUG, message)
	}
}

func PluginInfo(pluginName, message string) {
	logger := GetPluginLogger(pluginName)
	if logger != nil {
		logMessageWithLogger(logger, LEVEL_INFO, message)
	}
}

func PluginWarn(pluginName, message string) {
	logger := GetPluginLogger(pluginName)
	if logger != nil {
		logMessageWithLogger(logger, LEVEL_WARN, message)
	}
}

func PluginError(pluginName, message string) {
	logger := GetPluginLogger(pluginName)
	if logger != nil {
		logMessageWithLogger(logger, LEVEL_ERROR, message)
	}
}

func logMessageWithLogger(logger *Logger, level Level, message string) {
	if logger == nil || level < logger.level {
		return
	}

	if err := rotatePluginLogFile(logger); err != nil {
		log.Printf("Failed to rotate plugin log file: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), message)

	logger.logger.Print(formattedMessage)
}

func rotatePluginLogFile(logger *Logger) error {
	if logger == nil || logger.file == nil {
		return nil
	}

	stat, err := logger.file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() < MaxLogFileSize {
		return nil
	}

	logger.file.Close()

	// Find the plugin name from the loggers map
	var pluginName string
	for name, l := range pluginLoggers {
		if l == logger {
			pluginName = name
			break
		}
	}

	if pluginName == "" {
		return nil
	}

	for i := MaxLogFiles - 1; i >= 1; i-- {
		oldFile := filepath.Join(logger.logDir, fmt.Sprintf("%s-%s.log.%d", pluginName, time.Now().Format("2006-01-02"), i))
		newFile := filepath.Join(logger.logDir, fmt.Sprintf("%s-%s.log.%d", pluginName, time.Now().Format("2006-01-02"), i+1))

		if _, err := os.Stat(oldFile); err == nil {
			os.Rename(oldFile, newFile)
		}
	}

	currentFile := filepath.Join(logger.logDir, fmt.Sprintf("%s-%s.log", pluginName, time.Now().Format("2006-01-02")))
	rotatedFile := filepath.Join(logger.logDir, fmt.Sprintf("%s-%s.log.1", pluginName, time.Now().Format("2006-01-02")))
	os.Rename(currentFile, rotatedFile)

	file, err := os.OpenFile(currentFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logger.file = file
	logger.logger = log.New(file, "", 0)

	return nil
}

func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}

	for _, logger := range pluginLoggers {
		if logger != nil && logger.file != nil {
			logger.file.Close()
		}
	}
}
