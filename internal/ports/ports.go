package ports

import (
	"fmt"
	"os"
	"runtime"

	"github.com/abhimanyu003/pttr/internal/common"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadPorts() tea.Msg {
	defer func() {
		if r := recover(); r != nil {

			fmt.Fprintf(os.Stderr, "Panic in loadPorts: %v\n", r)
		}
	}()

	ports, err := GetOpenPorts()
	if err != nil {
		return common.PortsLoadedMsg{}
	}
	return common.PortsLoadedMsg(ports)
}

func GetOpenPorts() ([]common.PortInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsPorts()
	case "darwin":
		return getDarwinPorts()
	case "linux":
		return getLinuxPorts()
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return getBSDPorts()
	case "solaris":
		return getSolarisPorts()
	default:
		return getGenericUnixPorts()
	}
}
