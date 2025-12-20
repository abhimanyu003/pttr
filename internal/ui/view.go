package ui

import (
	"fmt"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func (m Model) View() string {
	if m.Loading && ((m.Mode == common.PortsMode && len(m.Ports) == 0) || (m.Mode == common.ProcessesMode && len(m.Processes) == 0) || (m.Mode == common.IPsMode && len(m.IPs) == 0)) {
		loadingText := "Loading ports..."
		if m.Mode == common.ProcessesMode {
			loadingText = "Loading processes..."
		} else if m.Mode == common.IPsMode {
			loadingText = "Loading network interfaces..."
		}
		return fmt.Sprintf("⏳ %s", loadingText)
	}
	var content strings.Builder

	content.WriteString(m.renderHeader())

	// Show notification if active
	if m.Notification != "" {
		content.WriteString("\n")
		content.WriteString(m.renderNotification())
	}

	content.WriteString("\n")
	content.WriteString(m.List.View())
	statusLine := m.renderStatusAndHelp()
	if statusLine != "" {
		content.WriteString("\n")
		content.WriteString(statusLine)
	}

	return content.String()
}

func (m Model) renderHeader() string {
	var title string
	var count int

	if m.Mode == common.PortsMode {
		count = len(m.Ports)
		title = fmt.Sprintf("[Net] %d Open Ports", count)
	} else if m.Mode == common.ProcessesMode {
		count = len(m.Processes)
		sortInfo := ""
		switch m.SortMode {
		case common.SortByName:
			sortInfo = "Name"
		case common.SortByCPU:
			sortInfo = "CPU"
		case common.SortByMemory:
			sortInfo = "Memory"
		case common.SortByPID:
			sortInfo = "PID"
		}

		viewInfo := "List"
		if m.ShowTree {
			viewInfo = "Tree"
		}

		title = fmt.Sprintf("[Sys] %d Processes • %s • %s", count, viewInfo, sortInfo)
	} else if m.Mode == common.IPsMode {
		count = len(m.IPs)
		title = fmt.Sprintf("[IP] %d Network Interfaces", count)
	}

	var privilegeIndicator string
	if m.Mode == common.ProcessesMode {
		if m.IsRoot {
			privilegeIndicator = " " + RootStyle.Render("[Root]")
		} else {
			privilegeIndicator = " " + UserStyle.Render("[User]")
		}
	}

	return TitleStyle.Render(title + privilegeIndicator)
}

func (m Model) renderNotification() string {
	if m.Notification == "" {
		return ""
	}

	// Style notification based on content
	msgLower := strings.ToLower(m.Notification)
	var styledNotification string

	if strings.Contains(msgLower, "✅") || strings.Contains(msgLower, "successfully") {
		styledNotification = SuccessStyle.Render("🎉 " + m.Notification)
	} else if strings.Contains(msgLower, "❌") || strings.Contains(msgLower, "failed") || strings.Contains(msgLower, "error") {
		styledNotification = ErrorStyle.Render("⚠️  " + m.Notification)
	} else {
		styledNotification = InfoStyle.Render("ℹ️  " + m.Notification)
	}

	return styledNotification
}

func (m Model) renderStatusAndHelp() string {
	var sections []string
	if m.Message != "" {
		msgLower := strings.ToLower(m.Message)
		var icon string
		var styledMessage string

		if strings.Contains(msgLower, "error") || strings.Contains(msgLower, "failed") {
			icon = "[ERR]"
			styledMessage = ErrorStyle.Render(m.Message)
		} else if strings.Contains(msgLower, "warning") || strings.Contains(msgLower, "permission") {
			icon = "[WARN]"
			styledMessage = WarningStyle.Render(m.Message)
		} else if strings.Contains(msgLower, "success") || strings.Contains(msgLower, "killed") {
			icon = "[OK]"
			styledMessage = InfoStyle.Render(m.Message)
		} else if strings.Contains(msgLower, "found") {
			icon = ""
			styledMessage = InfoStyle.Render(m.Message)
		} else {
			icon = ""
			styledMessage = InfoStyle.Render(m.Message)
		}
		if icon != "" {
			sections = append(sections, fmt.Sprintf("%s %s", icon, styledMessage))
		} else {
			sections = append(sections, styledMessage)
		}
	}

	var controlParts []string

	controlParts = append(controlParts,
		HelpStyle.Render("↑/k up"),
		HelpStyle.Render("↓/j down"),
		HelpStyle.Render("/ filter"),
		HelpStyle.Render("q quit"),
	)

	controlParts = append(controlParts,
		KillStyle.Render("k kill"),
		HelpStyle.Render("r refresh"),
	)

	if m.Mode == common.PortsMode {
		controlParts = append(controlParts,
			ProcessStyle.Render("p processes"),
			ProcessStyle.Render("i ips"))
	} else if m.Mode == common.ProcessesMode {
		controlParts = append(controlParts,
			ProcessStyle.Render("p ports"),
			ProcessStyle.Render("i ips"),
			HelpStyle.Render("s sort"),
			HelpStyle.Render("g tree"),
		)
	} else if m.Mode == common.IPsMode {
		controlParts = append(controlParts,
			ProcessStyle.Render("p ports"))
	}

	helpLine := strings.Join(controlParts, HelpStyle.Render(" • "))
	sections = append(sections, helpLine)

	if m.Mode == common.ProcessesMode && !m.IsRoot && len(sections) == 1 {
		warning := WarningStyle.Render("[WARN] Some processes need sudo")
		sections = append(sections, HelpStyle.Render(warning))
	}

	return strings.Join(sections, "\n")
}
