package agent

import (
	"github.com/icphalanx/agent/hosts/linux"
	"github.com/icphalanx/agent/types"
)

func MakeLocalAgent() (types.Host, error) {
	return linux.Create()
}
