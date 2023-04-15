package logging

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)
type FileLogger struct {
	file          *os.File
    maxFileSize   int64
    logDirectory  string
    logFilePrefix string
	// The logChan is used to send log messages to the logger goroutine.
	logChan       chan logEntry
	// The quitChan is used to send a signal to the logger goroutine to stop.
	quitChan      chan bool
}

type logEntry struct {
    level string
    msg   string
    args  []interface{}
}

func New(logDirectory, logFilePrefix string, maxFileSize int64) (*FileLogger, error) {
    if err := os.MkdirAll(logDirectory, 0755); err != nil {
        return nil, err
    }

    logFile, err := createLogFile(logDirectory)
    if err != nil {
        return nil, err
    }

    fileLogger := FileLogger{
        file:          logFile,
        maxFileSize:   maxFileSize,
        logDirectory:  logDirectory,
        logFilePrefix: logFilePrefix,
		logChan:       make(chan logEntry, 100),
        quitChan:      make(chan bool),
    }
	go fileLogger.logHandler()

    return &fileLogger, nil
}

// implement ILogger interface
func (fl *FileLogger) Debug(msg string, args ...interface{}) {
	fl.log("DEBUG", msg, args...)
}

func (fl *FileLogger) Info(msg string, args ...interface{}) {
	fl.log("INFO", msg, args...)
}

func (fl *FileLogger) Warn(msg string, args ...interface{}) {
	fl.log("WARN", msg, args...)
}

func (fl *FileLogger) Error(msg string, args ...interface{}) {
	fl.log("ERROR", msg, args...)
}

// createLogFile creates a new log file in the given directory with the given prefix.
func createLogFile(logDirectory string) (*os.File, error) {
    logFilePrefix := "chux-datastore-log"
	timestamp := time.Now().Format("2006-01-02")
    fileName := fmt.Sprintf("%s-%s.log", logFilePrefix, timestamp)
    filePath := filepath.Join(logDirectory, fileName)
    return os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}

func (fl *FileLogger) log(level, msg string, args ...interface{}) {
    fl.logChan <- logEntry{level: level, msg: msg, args: args}
}

func (fl *FileLogger) needsRotation() bool {
    fileInfo, err := fl.file.Stat()
    if err != nil {
        return false
    }
    return fileInfo.Size() >= fl.maxFileSize
}

func (fl *FileLogger) rotate() {
    // Close the current log file
    fl.file.Close()

    // Zip the old log file
    oldFilePath := fl.file.Name()
    zipFilePath := oldFilePath + ".zip"
    if err := zipFile(oldFilePath, zipFilePath); err != nil {
        fmt.Printf("Error zipping old log file: %v\n", err)
    } else {
        os.Remove(oldFilePath)
    }

    // Create a new log file
    newLogFile, err := createLogFile(fl.logDirectory)
    if err != nil {
        fmt.Printf("Error creating new log file: %v\n", err)
        return
    }
    fl.file = newLogFile
}

func (fl *FileLogger) logHandler() {
    for {
        select {
        case entry := <-fl.logChan:
            // Check if the file size exceeds the maximum size
            if fl.needsRotation() {
                fl.rotate()
            }

            log.SetOutput(fl.file)
            log.Printf("%s: %s", entry.level, fmt.Sprintf(entry.msg, entry.args...))
        case <-fl.quitChan:
            return
        }
    }
}

func (fl *FileLogger) Close() {
    fl.quitChan <- true
    close(fl.logChan)
    fl.file.Close()
}


// zipFile zips a log file and returns an error if any
// occurs during the process.
func zipFile(source, target string) error {
    zipFile, err := os.Create(target)
    if err != nil {
        return err
    }
    defer zipFile.Close()

    archive := zip.NewWriter(zipFile)
    defer archive.Close()

    sourceFile, err := os.Open(source)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    fileInfo, err := sourceFile.Stat()
    if err != nil {
        return err
    }

    header, err := zip.FileInfoHeader(fileInfo)
    if err != nil {
        return err
    }

    // Use the file name as the zip entry name
    header.Name = filepath.Base(source)

    // Set the compression method to Deflate
    header.Method = zip.Deflate

    writer, err := archive.CreateHeader(header)
    if err != nil {
        return err
    }

    _, err = io.Copy(writer, sourceFile)
    return err
}