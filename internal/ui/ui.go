package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"
)

var (
	logFile *os.File
	verbose bool
)

// Init initializes the UI logger.
// verbose: if true, enables Info, Success, and Step output to stdout.
func Init(configDir string, v bool) error {
	verbose = v
	
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0700)
	}

	logPath := filepath.Join(configDir, "debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

// SetVerbose allows changing verbosity at runtime
func SetVerbose(v bool) {
	verbose = v
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// PrintTable prints a formatted table to stdout
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	// Print headers
	for i, h := range headers {
		fmt.Fprintf(w, "%s\t", h)
		if i == len(headers)-1 {
			fmt.Fprintln(w)
		}
	}
	
	// Print separator (rough approximation based on header length is hard, just standard line)
	// We'll just rely on tabwriter alignment for clean look
	
	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Fprintf(w, "%s\t", cell)
			if i == len(row)-1 {
				fmt.Fprintln(w)
			}
		}
	}
	w.Flush()
}

func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fileLog("SUCCESS: " + msg)
	if verbose {
		fmt.Printf("[+] %s\n", msg)
	}
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fileLog("INFO: " + msg)
	if verbose {
		fmt.Printf("[*] %s\n", msg)
	}
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fileLog("ERROR: " + msg)
	fmt.Fprintf(os.Stderr, "[!] %s\n", msg)
}

func Step(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fileLog("STEP: " + msg)
	if verbose {
		fmt.Printf("    - %s\n", msg)
	}
}

func Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fileLog("DEBUG: " + msg)
	if verbose {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

func fileLog(msg string) {
	if logFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		io.WriteString(logFile, fmt.Sprintf("[%s] %s\n", timestamp, msg))
	}
}