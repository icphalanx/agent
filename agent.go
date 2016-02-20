package agent

import (
	"github.com/icphalanx/agent/hosts/linux"
	"github.com/icphalanx/agent/types"
)

func Run() (types.Host, error) {
	return linux.Create()
}
