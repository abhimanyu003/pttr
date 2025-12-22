package process

import (
	"path/filepath"
	"strings"

	"github.com/abhimanyu003/pttr/internal/common"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func EnhanceProcessInfo(processes []common.ProcessInfo) []common.ProcessInfo {
	enhanced := make([]common.ProcessInfo, len(processes))

	for i, proc := range processes {
		enhanced[i] = proc
		enhanced[i].CPUFloat = common.ParseFloat(strings.TrimSuffix(proc.CPU, "%"))
		enhanced[i].MemoryFloat = common.ParseFloat(strings.TrimSuffix(proc.Memory, "%"))
		enhanced[i].DisplayName = getBetterProcessName(proc.Process, proc.Command)
		enhanced[i].Icon = getSimpleProcessIcon(proc.Process)
	}

	return enhanced
}

func normalizeProcessName(name string) string {
	n := strings.ToLower(name)
	n = strings.TrimSuffix(n, ".exe")
	return n
}

func getBetterProcessName(process, command string) string {
	proc := normalizeProcessName(process)

	nameMap := map[string]string{
		// Browsers
		"chrome":  "Google Chrome",
		"msedge":  "Microsoft Edge",
		"firefox": "Mozilla Firefox",
		"safari":  "Safari",
		"brave":   "Brave Browser",
		"opera":   "Opera",

		// Editors / IDEs
		"code":         "Visual Studio Code",
		"code-helper":  "VS Code Helper",
		"idea":         "IntelliJ IDEA",
		"goland":       "GoLand",
		"pycharm":      "PyCharm",
		"webstorm":     "WebStorm",
		"sublime_text": "Sublime Text",
		"atom":         "Atom Editor",
		"notepad":      "Notepad",
		"notepad++":    "Notepad++",
		"vim":          "Vim",
		"emacs":        "Emacs",
		"gedit":        "Gedit",

		// Office / Productivity
		"winword":  "Microsoft Word",
		"excel":    "Microsoft Excel",
		"outlook":  "Microsoft Outlook",
		"textedit": "TextEdit",
		"preview":  "Preview",
		"mail":     "Mail",
		"calendar": "Calendar",

		// Languages / Runtimes
		"node":    "Node.js",
		"deno":    "Deno",
		"bun":     "Bun",
		"python":  "Python",
		"python3": "Python 3",
		"java":    "Java",
		"go":      "Go",
		"dotnet":  ".NET Runtime",
		"gcc":     "GCC Compiler",
		"make":    "Make",
		"git":     "Git",

		// Containers / Cloud
		"docker":     "Docker CLI",
		"dockerd":    "Docker Daemon",
		"containerd": "Containerd",
		"kubectl":    "Kubernetes CLI",
		"kubelet":    "Kubernetes Node Agent",
		"minikube":   "Minikube",
		"kind":       "Kubernetes in Docker",

		// Databases
		"mysql":        "MySQL",
		"mysqld":       "MySQL Server",
		"postgres":     "PostgreSQL",
		"redis-server": "Redis Server",
		"mongod":       "MongoDB",
		"etcd":         "etcd",
		"influxd":      "InfluxDB",

		// Web Servers
		"nginx":   "Nginx",
		"httpd":   "Apache HTTP Server",
		"apache2": "Apache",
		"caddy":   "Caddy Server",

		// Networking
		"ssh":       "SSH Client",
		"sshd":      "SSH Server",
		"openvpn":   "OpenVPN",
		"wireguard": "WireGuard",
		"curl":      "cURL",
		"wget":      "wget",

		// macOS system
		"kernel_task":      "macOS Kernel",
		"launchd":          "Launch Daemon",
		"windowserver":     "Window Server",
		"finder":           "Finder",
		"dock":             "macOS Dock",
		"activity monitor": "Activity Monitor",

		// Linux system
		"systemd":     "System Manager",
		"journald":    "System Journal",
		"cron":        "Cron Scheduler",
		"dbus-daemon": "D-Bus Daemon",
		"gnome-shell": "GNOME Shell",
		"plasma":      "KDE Plasma",

		// Windows system
		"explorer":   "Windows Explorer",
		"svchost":    "Windows Service Host",
		"lsass":      "Local Security Authority",
		"winlogon":   "Windows Logon",
		"services":   "Windows Services Manager",
		"taskmgr":    "Task Manager",
		"powershell": "PowerShell",
		"pwsh":       "PowerShell Core",
		"cmd":        "Command Prompt",
		"onedrive":   "OneDrive",

		// Communication
		"slack":    "Slack",
		"discord":  "Discord",
		"zoom":     "Zoom",
		"teams":    "Microsoft Teams",
		"telegram": "Telegram",
		"signal":   "Signal",

		// Media / Gaming
		"spotify": "Spotify",
		"vlc":     "VLC Media Player",
		"steam":   "Steam",
		"gimp":    "GIMP",
		"blender": "Blender",
		"ffmpeg":  "FFmpeg",

		// Other
		"inkscape": "Inkscape",
	}

	if better, ok := nameMap[proc]; ok {
		return better
	}

	// Partial match fallback
	for key, value := range nameMap {
		if strings.Contains(proc, key) {
			return value
		}
	}

	// macOS .app detection from command
	if strings.Contains(command, ".app") {
		parts := strings.SplitSeq(command, "/")
		for part := range parts {
			if before, ok := strings.CutSuffix(part, ".app"); ok {
				return before
			}
		}
	}

	// Use executable name from command path
	if command != "" {
		base := filepath.Base(command)
		base = normalizeProcessName(base)
		if base != "" && base != proc {
			return cases.Title(language.Und, cases.NoLower).String(base)
		}
	}
	if proc != "" {
		return cases.Title(language.Und, cases.NoLower).String(proc)
	}

	return process
}

func getSimpleProcessIcon(process string) string {
	proc := normalizeProcessName(process)

	iconMap := map[string]string{
		// Browsers
		"chrome":  "[Web]",
		"msedge":  "[Web]",
		"firefox": "[Web]",
		"safari":  "[Web]",
		"brave":   "[Web]",
		"opera":   "[Web]",

		// Dev / Editors
		"code":         "[Dev]",
		"idea":         "[Dev]",
		"goland":       "[Dev]",
		"node":         "[Dev]",
		"python":       "[Dev]",
		"java":         "[Dev]",
		"go":           "[Dev]",
		"vim":          "[Dev]",
		"emacs":        "[Dev]",
		"notepad":      "[Edit]",
		"sublime_text": "[Edit]",
		"gedit":        "[Edit]",

		// Office
		"winword":  "[Office]",
		"excel":    "[Office]",
		"outlook":  "[Office]",
		"textedit": "[Edit]",
		"preview":  "[View]",

		// Containers / K8s
		"docker":     "[Sys]",
		"dockerd":    "[Sys]",
		"containerd": "[Sys]",
		"kubectl":    "[K8s]",
		"kubelet":    "[K8s]",

		// Databases
		"mysql":        "[DB]",
		"mysqld":       "[DB]",
		"postgres":     "[DB]",
		"redis-server": "[DB]",
		"mongod":       "[DB]",
		"etcd":         "[DB]",

		// Web servers
		"nginx":   "[Web]",
		"httpd":   "[Web]",
		"apache2": "[Web]",
		"caddy":   "[Web]",

		// Terminal / Shell / Tools
		"terminal":   "[Term]",
		"iterm2":     "[Term]",
		"bash":       "[Term]",
		"zsh":        "[Term]",
		"fish":       "[Term]",
		"powershell": "[Term]",
		"pwsh":       "[Term]",
		"cmd":        "[Term]",
		"git":        "[Dev]",
		"gcc":        "[Dev]",
		"make":       "[Dev]",
		"curl":       "[Net]",
		"wget":       "[Net]",

		// System
		"systemd":     "[Sys]",
		"launchd":     "[Sys]",
		"kernel_task": "[Sys]",
		"svchost":     "[Sys]",
		"explorer":    "[Sys]",
		"gnome-shell": "[Sys]",
		"plasma":      "[Sys]",

		// Chat / Comm
		"slack":    "[Chat]",
		"discord":  "[Chat]",
		"teams":    "[Chat]",
		"zoom":     "[Chat]",
		"telegram": "[Chat]",
		"signal":   "[Chat]",

		// Media / Gaming
		"spotify":  "[Media]",
		"vlc":      "[Media]",
		"steam":    "[Game]",
		"gimp":     "[Media]",
		"blender":  "[Media]",
		"ffmpeg":   "[Media]",
		"inkscape": "[Media]",

		// Cloud
		"onedrive": "[Cloud]",
	}

	if icon, ok := iconMap[proc]; ok {
		return icon
	}

	// Heuristics
	switch {
	case strings.Contains(proc, "docker"):
		return "[Sys]"
	case strings.Contains(proc, "kube"):
		return "[K8s]"
	case strings.Contains(proc, "sql"),
		strings.Contains(proc, "mongo"),
		strings.Contains(proc, "redis"):
		return "[DB]"
	case strings.Contains(proc, "server"):
		return "[Srv]"
	case strings.HasSuffix(proc, "d"):
		return "[Daemon]"
	case strings.Contains(proc, "word") || strings.Contains(proc, "excel") || strings.Contains(proc, "outlook"):
		return "[Office]"
	case strings.Contains(proc, "steam"):
		return "[Game]"
	}

	return "[App]"
}
