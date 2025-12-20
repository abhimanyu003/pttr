package process

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadProcesses() tea.Msg {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Panic in loadProcesses: %v\n", r)
		}
	}()

	processes, err := GetAllProcesses()
	if err != nil {
		return common.ProcessesLoadedMsg{}
	}
	enhanced := EnhanceProcessInfo(processes)
	return common.ProcessesLoadedMsg(enhanced)
}

func GetAllProcesses() ([]common.ProcessInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsProcesses()
	case "darwin":
		return getDarwinProcesses()
	case "linux":
		return getLinuxProcesses()
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return getBSDProcesses()
	case "solaris":
		return getSolarisProcesses()
	default:
		return getGenericUnixProcesses()
	}
}

func SortProcesses(processes []common.ProcessInfo, sortMode common.SortMode) []common.ProcessInfo {
	sorted := make([]common.ProcessInfo, len(processes))
	copy(sorted, processes)

	switch sortMode {
	case common.SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].DisplayName) < strings.ToLower(sorted[j].DisplayName)
		})
	case common.SortByCPU:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CPUFloat > sorted[j].CPUFloat
		})
	case common.SortByMemory:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].MemoryFloat > sorted[j].MemoryFloat
		})
	case common.SortByPID:
		sort.Slice(sorted, func(i, j int) bool {
			pidI, _ := strconv.Atoi(sorted[i].PID)
			pidJ, _ := strconv.Atoi(sorted[j].PID)
			return pidI < pidJ
		})
	}

	return sorted
}

func GroupProcessesByParent(processes []common.ProcessInfo, sortMode common.SortMode) []common.ProcessInfo {
	processesWithPPID := addParentPIDs(processes)
	processMap := make(map[string]common.ProcessInfo)
	for _, proc := range processesWithPPID {
		processMap[proc.PID] = proc
	}
	children := make(map[string][]common.ProcessInfo)
	var roots []common.ProcessInfo
	for _, proc := range processesWithPPID {
		if proc.PPID == "" || proc.PPID == "0" {
			roots = append(roots, proc)
		} else {
			children[proc.PPID] = append(children[proc.PPID], proc)
		}
	}

	var result []common.ProcessInfo

	var addProcessAndChildren func(proc common.ProcessInfo, depth int, isLast bool, parentPrefix string)
	addProcessAndChildren = func(proc common.ProcessInfo, depth int, isLast bool, parentPrefix string) {

		treeProcess := proc
		var currentPrefix string
		if depth == 0 {
			currentPrefix = ""
		} else {
			var treeSymbol string
			if isLast {
				treeSymbol = "└── "
			} else {
				treeSymbol = "├── "
			}
			currentPrefix = parentPrefix + treeSymbol
		}

		if depth > 0 {
			treeProcess.DisplayName = currentPrefix + proc.DisplayName
		}
		result = append(result, treeProcess)
		if childProcs, exists := children[proc.PID]; exists {
			sortProcessSlice(childProcs, sortMode)
			var childPrefix string
			if depth == 0 {
				childPrefix = ""
			} else {
				if isLast {
					childPrefix = parentPrefix + "    "
				} else {
					childPrefix = parentPrefix + "│   "
				}
			}
			for i, child := range childProcs {
				isLastChild := i == len(childProcs)-1
				addProcessAndChildren(child, depth+1, isLastChild, childPrefix)
			}
		}
	}
	sortProcessSlice(roots, sortMode)
	for _, root := range roots {
		addProcessAndChildren(root, 0, true, "")
	}

	return result
}

func sortProcessSlice(processes []common.ProcessInfo, sortMode common.SortMode) {
	switch sortMode {
	case common.SortByName:
		sort.Slice(processes, func(i, j int) bool {
			return strings.ToLower(processes[i].DisplayName) < strings.ToLower(processes[j].DisplayName)
		})
	case common.SortByCPU:
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].CPUFloat > processes[j].CPUFloat
		})
	case common.SortByMemory:
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].MemoryFloat > processes[j].MemoryFloat
		})
	case common.SortByPID:
		sort.Slice(processes, func(i, j int) bool {
			pidI, _ := strconv.Atoi(processes[i].PID)
			pidJ, _ := strconv.Atoi(processes[j].PID)
			return pidI < pidJ
		})
	}
}
