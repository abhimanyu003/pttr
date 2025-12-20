package ui

import (
	"time"

	"github.com/abhimanyu003/pttr/internal/common"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type NotificationMsg struct{}

type Model struct {
	List             list.Model
	Ports            []common.PortInfo
	Processes        []common.ProcessInfo
	IPs              []common.IPInfo
	Selected         int
	Loading          bool
	Message          string
	Mode             common.ViewMode
	SortMode         common.SortMode
	ShowTree         bool
	IsRoot           bool
	Username         string
	PendingUpdate    bool
	Notification     string
	NotificationTime time.Time
}

func NewModel(startMode common.ViewMode) Model {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 80, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	return Model{
		List:      l,
		Ports:     []common.PortInfo{},
		Processes: []common.ProcessInfo{},
		IPs:       []common.IPInfo{},
		Selected:  0,
		Loading:   true,
		Mode:      startMode,
		SortMode:  common.SortByCPU,
		ShowTree:  false,
		IsRoot:    common.IsRunningAsRoot(),
		Username:  common.GetCurrentUser(),
	}
}

// Helper function to create a notification that auto-clears after 3 seconds
func (m *Model) SetNotification(message string) tea.Cmd {
	m.Notification = message
	m.NotificationTime = time.Now()
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return NotificationMsg{}
	})
}
