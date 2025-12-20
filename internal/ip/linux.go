package ip

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getLinuxIPs() ([]common.IPInfo, error) {
	cmd := exec.Command("ip", "addr", "show")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("ifconfig", "-a")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("both ip and ifconfig commands failed: %v", err)
		}
		return parseUnixIfconfig(string(output)), nil
	}

	return parseLinuxIPAddr(string(output)), nil
}

func parseLinuxIPAddr(output string) []common.IPInfo {
	var ips []common.IPInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentInterface string
	var currentMAC string
	var currentStatus string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, ":") && (strings.Contains(line, "<") || strings.Contains(line, "mtu")) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentInterface = strings.TrimSuffix(parts[1], ":")
				if strings.Contains(line, "UP") && !strings.Contains(line, "DOWN") {
					currentStatus = "UP"
				} else {
					currentStatus = "DOWN"
				}
			}
		}

		if strings.Contains(line, "link/ether") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentMAC = parts[1]
			}
		}

		if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ipWithMask := parts[1]
				ipAddr := strings.Split(ipWithMask, "/")[0]
				var netmask, broadcast string
				if strings.Contains(ipWithMask, "/") {
					maskBits := strings.Split(ipWithMask, "/")[1]
					netmask = cidrToNetmask(maskBits)
				}
				for i, part := range parts {
					if part == "brd" && i+1 < len(parts) {
						broadcast = parts[i+1]
						break
					}
				}

				ips = append(ips, common.IPInfo{
					Interface: currentInterface,
					IPAddress: ipAddr,
					Type:      "IPv4",
					Status:    currentStatus,
					MAC:       currentMAC,
					Netmask:   netmask,
					Broadcast: broadcast,
				})
			}
		}

		if strings.Contains(line, "inet6 ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ipWithMask := parts[1]

				ipAddr := strings.Split(ipWithMask, "/")[0]

				ips = append(ips, common.IPInfo{
					Interface: currentInterface,
					IPAddress: ipAddr,
					Type:      "IPv6",
					Status:    currentStatus,
					MAC:       currentMAC,
				})
			}
		}
	}

	return ips
}

func cidrToNetmask(cidr string) string {
	bits, err := strconv.Atoi(cidr)
	if err != nil || bits < 0 || bits > 32 {
		return ""
	}

	mask := uint32(0xFFFFFFFF << (32 - bits))

	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF,
		(mask>>16)&0xFF,
		(mask>>8)&0xFF,
		mask&0xFF)
}
