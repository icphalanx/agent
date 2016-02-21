package syslogsocket

import (
	"fmt"
	"github.com/icphalanx/agent/reporters"
	"github.com/icphalanx/agent/types"
	"net"
	"os"
)

func init() {
	reporters.Register(SyslogSocketReporterFactory{})
}

type SyslogSocketReporterFactory struct{}

func (SyslogSocketReporterFactory) Id() string {
	return "syslogsocket"
}

func (ssrf SyslogSocketReporterFactory) Create(h types.Host) (types.Reporter, error) {
	relevant, err := ssrf.ApplicableTo(h)
	if err != nil {
		return nil, err
	}
	if !relevant {
		return nil, fmt.Errorf("SyslogSocketReporter not relevant for this system")
	}

	os.Remove("/run/systemd/journal/syslog")

	ua, err := net.ResolveUnixAddr("unixgram", "/run/systemd/journal/syslog")
	if err != nil {
		return nil, err
	}

	ln, err := net.ListenUnixgram("unixgram", ua)
	if err != nil {
		return nil, err
	}

	return &SyslogSocketReporter{
		host: h,
		conn: ln,
	}, nil
}

func (SyslogSocketReporterFactory) ApplicableTo(h types.Host) (bool, error) {
	// is this a LinuxHost?
	if !h.IsLocal() {
		return false, nil
	}

	_, err := os.Stat("/run/systemd/journal")
	if err != nil {
		return false, nil
	}

	return true, nil
}
