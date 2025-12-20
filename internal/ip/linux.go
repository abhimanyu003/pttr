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
	// Try ip command first (modern Linux)
	cmd := exec.Command("ip", "addr", "show")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		ips := parseLinuxIPAddr(string(output))
		if len(ips) > 0 {
			return ips, nil
		}
	}

	// Try ip with different flags
	cmd = exec.Command("ip", "a")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		ips := parseLinuxIPAddr(string(output))
		if len(ips) > 0 {
			return ips, nil
		}
	}

	// Fallback to ifconfig
	cmd = exec.Command("ifconfig", "-a")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		ips := parseUnixIfconfig(string(output))
		if len(ips) > 0 {
			return ips, nil
		}
	}

	// Try ifconfig without -a flag
	cmd = exec.Command("ifconfig")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		ips := parseUnixIfconfig(string(output))
		if len(ips) > 0 {
			return ips, nil
		}
	}

	// Last resort: try to read /proc/net/dev for interface names and combine with hostname -I
	return getLinuxIPsFallback()
}

func getLinuxIPsFallback() ([]common.IPInfo, error) {
	var ips []common.IPInfo

	// Try to get interface names from /proc/net/dev
	cmd := exec.Command("cat", "/proc/net/dev")
	output, err := cmd.Output()
	if err == nil {
		interfaces := parseLinuxInterfaces(string(output))

		// For each interface, try to get its IP
		for _, iface := range interfaces {
			cmd := exec.Command("ip", "addr", "show", iface)
			output, err := cmd.Output()
			if err == nil {
				ifaceIPs := parseLinuxIPAddr(string(output))
				ips = append(ips, ifaceIPs...)
			}
		}
	}

	if len(ips) > 0 {
		return ips, nil
	}

	return nil, fmt.Errorf("all network interface detection methods failed (ip, ifconfig, /proc/net/dev)")
}

func parseLinuxInterfaces(output string) []string {
	var interfaces []string
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "Inter-") && !strings.HasPrefix(line, "face") {
			parts := strings.Split(line, ":")
			if len(parts) >= 1 {
				iface := strings.TrimSpace(parts[0])
				if iface != "" && iface != "lo" { // Skip loopback for this fallback
					interfaces = append(interfaces, iface)
				}
			}
		}
	}

	return interfaces
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

		// Parse interface line (e.g., "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500")
		if strings.Contains(line, ":") && (strings.Contains(line, "<") || strings.Contains(line, "mtu")) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Handle both "2: eth0:" and "eth0:" formats
				interfacePart := parts[1]
				if strings.Contains(parts[0], ":") && !strings.Contains(parts[1], ":") {
					// Format: "eth0 <flags>"
					interfacePart = parts[0]
				}
				currentInterface = strings.TrimSuffix(interfacePart, ":")

				// Determine status from flags
				if strings.Contains(line, "UP") && !strings.Contains(line, "DOWN") {
					currentStatus = "UP"
				} else {
					currentStatus = "DOWN"
				}
			}
		}

		// Parse MAC address line (e.g., "link/ether 00:11:22:33:44:55 brd ff:ff:ff:ff:ff:ff")
		if strings.Contains(line, "link/ether") || strings.Contains(line, "link/loopback") {
			parts := strings.Fields(line)
			if len(parts) >= 2 && strings.Contains(line, "link/ether") {
				currentMAC = parts[1]
			} else if strings.Contains(line, "link/loopback") {
				currentMAC = "00:00:00:00:00:00"
			}
		}

		// Parse IPv4 address line (e.g., "inet 192.168.1.100/24 brd 192.168.1.255 scope global eth0")
		if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ipWithMask := parts[1]
				ipAddr := strings.Split(ipWithMask, "/")[0]
				var netmask, broadcast string

				// Convert CIDR to netmask
				if strings.Contains(ipWithMask, "/") {
					maskBits := strings.Split(ipWithMask, "/")[1]
					netmask = cidrToNetmask(maskBits)
				}

				// Find broadcast address
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

		// Parse IPv6 address line (e.g., "inet6 fe80::1/64 scope link")
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
