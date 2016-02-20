package linux

import (
	"github.com/icphalanx/agent/reporters"
	_ "github.com/icphalanx/agent/reporters/packagekit"
	"github.com/icphalanx/agent/types"
	"os"
)

type LinuxHost struct {
	reporters []types.Reporter
}

func (LinuxHost) Id() string {
	return "linux"
}

func (LinuxHost) IsLocal() bool {
	return true
}

func (LinuxHost) HumanName() (string, error) {
	return os.Hostname()
}

func (LinuxHost) Parent() (types.Host, error) {
	return nil, nil
}

func (lh *LinuxHost) Reporters() ([]types.Reporter, error) {
	return lh.reporters, nil
}

func Create() (types.Host, error) {
	lh := new(LinuxHost)
	lh.reporters = reporters.GenerateFor(lh)

	return lh, nil
}
