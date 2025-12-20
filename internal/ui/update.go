package ui

import (
	"fmt"
	"time"

	"github.com/abhimanyu003/pttr/internal/common"
	"github.com/abhimanyu003/pttr/internal/ip"
	"github.com/abhimanyu003/pttr/internal/ports"
	"github.com/abhimanyu003/pttr/internal/process"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	if m.Mode == common.ProcessesMode {
		return tea.Batch(
			process.LoadProcesses,
			m.List.StartSpinner(),
		)
	} else if m.Mode == common.IPsMode {
		return tea.Batch(
			ip.LoadIPs,
			m.List.StartSpinner(),
		)
	} else {
		return tea.Batch(
			ports.LoadPorts,
			m.List.StartSpinner(),
		)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		listHeight := msg.Height - 4
		if listHeight < 8 {
			listHeight = 8
		}
		m.List.SetSize(msg.Width, listHeight)
		return m, nil

	case tea.KeyMsg:
		if m.List.FilterState() == list.Filtering {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.List, cmd = m.List.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t", "p":

			if m.Mode == common.PortsMode {
				m.Mode = common.ProcessesMode
				m.Loading = true
				m.Message = "Loading processes..."
				m.List.ResetSelected()
				return m, tea.Batch(process.LoadProcesses, m.List.StartSpinner())
			} else if m.Mode == common.ProcessesMode {
				m.Mode = common.PortsMode
				m.Loading = true
				m.Message = "Loading ports..."
				m.List.ResetSelected()
				return m, tea.Batch(ports.LoadPorts, m.List.StartSpinner())
			} else {
				m.Mode = common.PortsMode
				m.Loading = true
				m.Message = "Loading ports..."
				m.List.ResetSelected()
				return m, tea.Batch(ports.LoadPorts, m.List.StartSpinner())
			}
		case "i":

			m.Mode = common.IPsMode
			m.Loading = true
			m.Message = "Loading network interfaces..."
			m.List.ResetSelected()
			return m, tea.Batch(ip.LoadIPs, m.List.StartSpinner())
		case "o":

			if m.Mode == common.IPsMode {
				m.Mode = common.ProcessesMode
				m.Loading = true
				m.Message = "Loading processes..."
				m.List.ResetSelected()
				return m, tea.Batch(process.LoadProcesses, m.List.StartSpinner())
			}
		case "r":
			m.Loading = true
			m.List.ResetSelected()
			if m.Mode == common.PortsMode {
				m.Message = "Refreshing ports..."
				return m, tea.Batch(ports.LoadPorts, m.List.StartSpinner())
			} else if m.Mode == common.ProcessesMode {
				m.Message = "Refreshing processes..."
				return m, tea.Batch(process.LoadProcesses, m.List.StartSpinner())
			} else if m.Mode == common.IPsMode {
				m.Message = "Refreshing network interfaces..."
				return m, tea.Batch(ip.LoadIPs, m.List.StartSpinner())
			}
		case "s":

			if m.Mode == common.ProcessesMode {

				switch m.SortMode {
				case common.SortByCPU:
					m.SortMode = common.SortByMemory
					m.Message = "Sorting by Memory usage"
				case common.SortByMemory:
					m.SortMode = common.SortByName
					m.Message = "Sorting by Name"
				case common.SortByName:
					m.SortMode = common.SortByPID
					m.Message = "Sorting by PID"
				case common.SortByPID:
					m.SortMode = common.SortByCPU
					m.Message = "Sorting by CPU usage"
				}
				if m.List.FilterValue() != "" {
					m.PendingUpdate = true
					m.Message += " (will apply when filter cleared)"
					return m, nil
				}

				if len(m.Processes) > 0 {
					var sortedProcesses []common.ProcessInfo
					if m.ShowTree {
						sortedProcesses = process.GroupProcessesByParent(m.Processes, m.SortMode)
					} else {
						sortedProcesses = process.SortProcesses(m.Processes, m.SortMode)
					}

					items := make([]list.Item, len(sortedProcesses))
					for i, proc := range sortedProcesses {
						items[i] = proc
					}

					m.List.SetItems(items)
					m.PendingUpdate = false
				}
				return m, nil
			}
		case "g":
			if m.Mode == common.ProcessesMode {
				m.ShowTree = !m.ShowTree
				if m.ShowTree {
					m.Message = "Showing process tree"
				} else {
					m.Message = "Showing flat process list"
				}

				if m.List.FilterValue() != "" {
					m.PendingUpdate = true
					m.Message += " (will apply when filter cleared)"
					return m, nil
				}

				if len(m.Processes) > 0 {
					var organizedProcesses []common.ProcessInfo
					if m.ShowTree {
						organizedProcesses = process.GroupProcessesByParent(m.Processes, m.SortMode)
					} else {
						organizedProcesses = process.SortProcesses(m.Processes, m.SortMode)
					}

					items := make([]list.Item, len(organizedProcesses))
					for i, proc := range organizedProcesses {
						items[i] = proc
					}

					m.List.SetItems(items)
					m.PendingUpdate = false
				}
				return m, nil
			}
		case "enter", "k":
			if m.List.SelectedItem() != nil {
				var pid string
				var processName string

				if m.Mode == common.PortsMode && len(m.Ports) > 0 {
					if selected, ok := m.List.SelectedItem().(common.PortInfo); ok {
						pid = selected.PID
						processName = selected.Process
					}
				} else if m.Mode == common.ProcessesMode && len(m.Processes) > 0 {
					if selected, ok := m.List.SelectedItem().(common.ProcessInfo); ok {
						pid = selected.PID
						processName = selected.DisplayName
						if processName == "" {
							processName = selected.Process
						}
					}
				} else if m.Mode == common.IPsMode {
					m.Message = "❌ Cannot kill network interfaces"
					return m, nil
				}

				if pid != "" && pid != "-" {
					m.Message = fmt.Sprintf("⏳ Killing process '%s' (PID: %s)...", processName, pid)
					return m, common.KillProcessWithName(pid, processName)
				} else {
					m.Message = "❌ Cannot kill: No PID available"
				}
			}
		}

	case common.PortsLoadedMsg:
		m.Loading = false

		ports := make([]common.PortInfo, len(msg))
		copy(ports, []common.PortInfo(msg))
		m.Ports = ports

		items := make([]list.Item, len(m.Ports))
		for i, port := range m.Ports {
			items[i] = port
		}

		m.List.SetItems(items)

		if len(m.Ports) == 0 {
			m.Message = "No open ports found"
		} else {
			m.Message = fmt.Sprintf("Found %d open ports", len(m.Ports))
		}

		m.List.StopSpinner()
		return m, nil

	case common.ProcessesLoadedMsg:
		m.Loading = false

		processes := make([]common.ProcessInfo, len(msg))
		copy(processes, []common.ProcessInfo(msg))
		m.Processes = processes

		var organizedProcesses []common.ProcessInfo
		if m.ShowTree {
			organizedProcesses = process.GroupProcessesByParent(m.Processes, m.SortMode)
		} else {
			organizedProcesses = process.SortProcesses(m.Processes, m.SortMode)
		}

		items := make([]list.Item, len(organizedProcesses))
		for i, proc := range organizedProcesses {
			items[i] = proc
		}

		m.List.SetItems(items)

		if len(m.Processes) == 0 {
			m.Message = "No processes found"
		} else {
			m.Message = fmt.Sprintf("Found %d running processes", len(m.Processes))
		}

		m.List.StopSpinner()
		return m, nil

	case common.IPsLoadedMsg:
		m.Loading = false

		ips := make([]common.IPInfo, len(msg))
		copy(ips, []common.IPInfo(msg))
		m.IPs = ips

		items := make([]list.Item, len(m.IPs))
		for i, ip := range m.IPs {
			items[i] = ip
		}

		m.List.SetItems(items)

		if len(m.IPs) == 0 {
			m.Message = "No network interfaces found"
		} else {
			m.Message = fmt.Sprintf("Found %d network interfaces", len(m.IPs))
		}

		m.List.StopSpinner()
		return m, nil

	case common.KillResultMsg:
		m.Message = msg.Message
		var notificationCmd tea.Cmd

		if msg.Success {
			notificationCmd = m.SetNotification(msg.Message)
			m.List.ResetFilter()
			if m.Mode == common.PortsMode {
				return m, tea.Batch(ports.LoadPorts, m.List.StartSpinner(), notificationCmd)
			} else if m.Mode == common.ProcessesMode {
				return m, tea.Batch(process.LoadProcesses, m.List.StartSpinner(), notificationCmd)
			} else if m.Mode == common.IPsMode {
				return m, tea.Batch(ip.LoadIPs, m.List.StartSpinner(), notificationCmd)
			}
		} else {
			notificationCmd = m.SetNotification(msg.Message)
			return m, notificationCmd
		}
		return m, nil

	case NotificationMsg:
		if time.Since(m.NotificationTime) >= 3*time.Second {
			m.Notification = ""
		}
		return m, nil
	}

	var cmd tea.Cmd
	oldFilterValue := m.List.FilterValue()
	m.List, cmd = m.List.Update(msg)
	newFilterValue := m.List.FilterValue()
	if m.PendingUpdate && oldFilterValue != "" && newFilterValue == "" {
		if m.Mode == common.ProcessesMode && len(m.Processes) > 0 {
			var organizedProcesses []common.ProcessInfo
			if m.ShowTree {
				organizedProcesses = process.GroupProcessesByParent(m.Processes, m.SortMode)
			} else {
				organizedProcesses = process.SortProcesses(m.Processes, m.SortMode)
			}

			items := make([]list.Item, len(organizedProcesses))
			for i, proc := range organizedProcesses {
				items[i] = proc
			}

			m.List.SetItems(items)
			m.PendingUpdate = false
			m.Message = "Applied pending changes"
		}
	}

	return m, cmd
}
