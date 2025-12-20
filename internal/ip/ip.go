package ip

import (
	"fmt"
	"os"
	"runtime"

	"github.com/abhimanyu003/pttr/internal/common"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadIPs() tea.Msg {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Panic in loadIPs: %v\n", r)
		}
	}()

	ips, err := GetAllIPs()
	if err != nil {
		return common.IPsLoadedMsg{}
	}
	return common.IPsLoadedMsg(ips)
}

func GetAllIPs() ([]common.IPInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsIPs()
	case "darwin":
		return getDarwinIPs()
	case "linux":
		return getLinuxIPs()
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return getBSDIPs()
	case "solaris":
		return getSolarisIPs()
	default:
		return getGenericUnixIPs()
	}
}
