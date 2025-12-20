package ports

import (
	"fmt"
	"os/exec"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getLinuxPorts() ([]common.PortInfo, error) {
	cmd := exec.Command("lsof", "-i", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("netstat", "-tuln")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("both lsof and netstat failed: %v", err)
		}
		ports, err := parseNetstatOutput(string(output))
		return ports, err
	}

	return parseUnixLsof(string(output)), nil
}
