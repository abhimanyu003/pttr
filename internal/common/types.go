package common

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	PortsMode ViewMode = iota
	ProcessesMode
	IPsMode
)

// SortMode represents different sorting options for processes
type SortMode int

const (
	SortByName SortMode = iota
	SortByCPU
	SortByMemory
	SortByPID
)

// PortInfo represents information about an open port
type PortInfo struct {
	Port    string
	PID     string
	Process string
	Proto   string
	State   string
}

func (p PortInfo) FilterValue() string {
	return p.Port + " " + p.Process + " " + p.Proto
}

func (p PortInfo) Title() string {
	// Add protocol-specific icons
	var icon string
	switch p.Proto {
	case "TCP":
		icon = ""
	case "UDP":
		icon = ""
	default:
		icon = ""
	}

	title := icon + " Port " + p.Port + " (" + p.Proto + ")"
	return strings.TrimSpace(title)
}

func (p PortInfo) Description() string {
	var parts []string

	if p.PID != "" && p.PID != "-" {
		parts = append(parts, "PID: "+p.PID)
	}

	if p.Process != "" {
		parts = append(parts, "Process: "+p.Process)
	}

	if p.State != "" {
		var stateDisplay string
		switch p.State {
		case "LISTEN":
			stateDisplay = "LISTEN"
		case "ESTABLISHED":
			stateDisplay = "ESTABLISHED"
		default:
			stateDisplay = p.State
		}
		parts = append(parts, "State: "+stateDisplay)
	}

	return joinParts(parts)
}

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID         string
	Process     string
	CPU         string
	Memory      string
	Command     string
	PPID        string  // Parent Process ID
	CPUFloat    float64 // For sorting
	MemoryFloat float64 // For sorting
	Icon        string  // Process icon/indicator
	DisplayName string  // Better display name
}

func (p ProcessInfo) FilterValue() string {
	return p.PID + " " + p.Process + " " + p.Command
}

func (p ProcessInfo) Title() string {
	displayName := p.DisplayName
	if displayName == "" {
		displayName = p.Process
	}

	icon := p.Icon
	if icon == "" {
		icon = "[App]"
	}

	// Add CPU/Memory indicators for high usage
	var indicators []string
	if p.CPUFloat > 10.0 {
		indicators = append(indicators, "🔥") // High CPU
	}
	if p.MemoryFloat > 10.0 {
		indicators = append(indicators, "💾") // High Memory
	}

	indicatorStr := ""
	if len(indicators) > 0 {
		indicatorStr = " " + joinStrings(indicators, "")
	}

	return icon + " " + displayName + " (PID: " + p.PID + ")" + indicatorStr
}

func (p ProcessInfo) Description() string {
	var parts []string

	// Resource usage with compact visual indicators
	if p.CPU != "" && p.Memory != "" {
		cpuBar := GetUsageBar(p.CPUFloat, 8) // Smaller bars
		memBar := GetUsageBar(p.MemoryFloat, 8)
		parts = append(parts, "CPU: "+p.CPU+" "+cpuBar)
		parts = append(parts, "Mem: "+p.Memory+" "+memBar)
	}

	// Parent process info (compact)
	if p.PPID != "" && p.PPID != "0" && p.PPID != "1" {
		parts = append(parts, "Parent: "+p.PPID)
	}

	// Command (more compact)
	cmd := p.Command
	if len(cmd) > 50 {
		if len(cmd) >= 47 {
			cmd = cmd[:47] + "..."
		} else {
			cmd = cmd + "..."
		}
	}
	if cmd != "" {
		parts = append(parts, cmd)
	}

	return joinParts(parts)
}

// IPInfo represents information about a network interface
type IPInfo struct {
	Interface string
	IPAddress string
	Type      string // IPv4, IPv6
	Status    string // UP, DOWN
	MAC       string // MAC address
	Netmask   string // Network mask
	Broadcast string // Broadcast address
}

func (ip IPInfo) FilterValue() string {
	return ip.Interface + " " + ip.IPAddress + " " + ip.Type + " " + ip.Status
}

func (ip IPInfo) Title() string {
	// Add type-specific icons
	var icon string
	switch ip.Type {
	case "IPv4":
		icon = ""
	case "IPv6":
		icon = ""
	default:
		icon = ""
	}
	title := icon + " " + ip.IPAddress + " (" + ip.Type + ") - " + ip.Interface

	return strings.TrimSpace(title)
}

func (ip IPInfo) Description() string {
	var parts []string

	// Status first
	var statusDisplay string
	switch ip.Status {
	case "UP":
		statusDisplay = "UP"
	case "DOWN":
		statusDisplay = "DOWN"
	default:
		statusDisplay = ip.Status
	}
	parts = append(parts, statusDisplay)

	// MAC address
	if ip.MAC != "" && ip.MAC != "00:00:00:00:00:00" {
		parts = append(parts, "MAC: "+ip.MAC)
	}

	// Network info
	if ip.Netmask != "" {
		parts = append(parts, "Netmask: "+ip.Netmask)
	}

	if ip.Broadcast != "" && ip.Broadcast != "0.0.0.0" {
		parts = append(parts, "Broadcast: "+ip.Broadcast)
	}

	return joinParts(parts)
}

type PortsLoadedMsg []PortInfo
type ProcessesLoadedMsg []ProcessInfo
type IPsLoadedMsg []IPInfo
type KillResultMsg struct {
	Success bool
	Message string
}

// Ensure all types implement list.Item interface
var _ list.Item = PortInfo{}
var _ list.Item = ProcessInfo{}
var _ list.Item = IPInfo{}

// Helper functions
func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var result strings.Builder
	result.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		result.WriteString(" • " + parts[i])
	}
	return result.String()
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	var result strings.Builder
	result.WriteString(strs[0])
	for i := 1; i < len(strs); i++ {
		result.WriteString(sep + strs[i])
	}
	return result.String()
}
