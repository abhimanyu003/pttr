package common

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GetUsageBar creates a visual usage bar
func GetUsageBar(usage float64, width int) string {
	if usage <= 0 || width <= 0 {
		if width <= 0 {
			return ""
		}
		return strings.Repeat("░", width)
	}

	filled := min(int((usage/100.0)*float64(width)), width)

	var bar strings.Builder
	for i := range width {
		if i < filled {
			if usage > 80 {
				bar.WriteString("█") // High usage - red
			} else if usage > 50 {
				bar.WriteString("▓") // Medium usage - yellow
			} else {
				bar.WriteString("▒") // Low usage - green
			}
		} else {
			bar.WriteString("░") // Empty
		}
	}

	return bar.String()
}

// ParseFloat safely parses a float from string
func ParseFloat(s string) float64 {
	if s == "" {
		return 0.0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return val
}

// IsCommandAvailable checks if a command exists in PATH
func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// GetCurrentUser returns the current user information
func GetCurrentUser() string {
	switch runtime.GOOS {
	case "windows":
		if username := os.Getenv("USERNAME"); username != "" {
			return username
		}
		return "unknown"
	default:
		if username := os.Getenv("USER"); username != "" {
			return username
		}
		if username := os.Getenv("LOGNAME"); username != "" {
			return username
		}
		return "unknown"
	}
}

// IsRunningAsRoot checks if the application is running with elevated privileges
func IsRunningAsRoot() bool {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	case "darwin", "linux", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
		cmd := exec.Command("id", "-u")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		uid := strings.TrimSpace(string(output))
		return uid == "0"
	default:
		return false
	}
}

// IsPermissionError checks if the error is related to insufficient permissions
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())

	permissionKeywords := []string{
		"permission denied",
		"access denied",
		"operation not permitted",
		"insufficient privileges",
		"access is denied",
		"not authorized",
		"eperm",
		"eacces",
	}

	for _, keyword := range permissionKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}
