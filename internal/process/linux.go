package process

import (
	"github.com/abhimanyu003/pttr/internal/common"
)

func getLinuxProcesses() ([]common.ProcessInfo, error) {
	return getDarwinProcesses()
}
