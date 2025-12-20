package ip

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
)

func getDarwinIPs() ([]common.IPInfo, error) {
	return parseDarwinIfconfig(), nil
}

func parseDarwinIfconfig() []common.IPInfo {
	var ips []common.IPInfo

	cmd := exec.Command("ifconfig", "-l")
	output, err := cmd.Output()
	if err != nil {
		return ips
	}

	interfaces := strings.Fields(strings.TrimSpace(string(output)))

	for _, interfaceName := range interfaces {
		cmd := exec.Command("ifconfig", interfaceName)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		var currentMAC string
		var currentStatus string = "DOWN"

		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "flags=") {
				if strings.Contains(line, "UP") {
					currentStatus = "UP"
				}
			}

			if strings.Contains(line, "ether ") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "ether" && i+1 < len(parts) {
						currentMAC = parts[i+1]
						break
					}
				}
			}

			if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
				parts := strings.Fields(line)
				var ipAddr, netmask, broadcast string
				for i, part := range parts {
					if part == "inet" && i+1 < len(parts) {
						ipAddr = parts[i+1]
					} else if part == "netmask" && i+1 < len(parts) {
						netmask = parts[i+1]
						if strings.HasPrefix(netmask, "0x") {
							netmask = hexNetmaskToDotted(netmask)
						}
					} else if part == "broadcast" && i+1 < len(parts) {
						broadcast = parts[i+1]
					}
				}

				if ipAddr != "" {
					ips = append(ips, common.IPInfo{
						Interface: interfaceName,
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
				for i, part := range parts {
					if part == "inet6" && i+1 < len(parts) {
						ipAddr := parts[i+1]
						if strings.Contains(ipAddr, "/") {
							ipAddr = strings.Split(ipAddr, "/")[0]
						}
						ips = append(ips, common.IPInfo{
							Interface: interfaceName,
							IPAddress: ipAddr,
							Type:      "IPv6",
							Status:    currentStatus,
							MAC:       currentMAC,
						})
						break
					}
				}
			}
		}
	}

	return ips
}

func hexNetmaskToDotted(hexMask string) string {
	if !strings.HasPrefix(hexMask, "0x") {
		return hexMask
	}

	hexStr := strings.TrimPrefix(hexMask, "0x")
	if len(hexStr) != 8 {
		return hexMask
	}

	var bytes [4]int64
	for i := 0; i < 4; i++ {
		byteStr := hexStr[i*2 : (i+1)*2]
		val, err := strconv.ParseInt(byteStr, 16, 64)
		if err != nil {
			return hexMask
		}
		bytes[i] = val
	}

	return fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3])
}
