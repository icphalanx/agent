package syslogsocket

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/icphalanx/agent/types"
	"github.com/jeromer/syslogparser/rfc3164"
)

type SyslogSocketReporter struct {
	host types.Host
	conn net.Conn
}

func (SyslogSocketReporter) Id() string {
	return "syslogsocket"
}

func (SyslogSocketReporter) Issues() ([]types.Issue, error) {
	return []types.Issue{}, nil
}

func (SyslogSocketReporter) Metrics() ([]types.Metric, error) {
	return []types.Metric{}, nil
}

func (SyslogSocketReporter) Hosts() ([]types.Host, error) {
	return []types.Host{}, nil
}

func (ssr *SyslogSocketReporter) LogLines() <-chan types.ReporterLogLine {
	s := make(chan types.ReporterLogLine)
	buf := make([]byte, 4096)
	go func() {
		for {
			n, err := ssr.conn.Read(buf)
			if err != nil {
				log.Println(":(", err)
				return
			}

			p := rfc3164.NewParser(buf[:n])
			if err := p.Parse(); err != nil {
				log.Println(":((", err)
				return
			}

			dmp := p.Dump()

			tags := make([]string, 3, 4)
			tags[0] = fmt.Sprintf("priority-%v", dmp["priority"])
			tags[1] = fmt.Sprintf("facility-%v", dmp["facility"])
			tags[2] = fmt.Sprintf("severity-%v", dmp["severity"])
			if dmp["tag"] != "" {
				tags = append(tags, fmt.Sprintf("tag-%v", dmp["tag"]))
			}

			s <- types.ReporterLogLine{
				Host:      ssr.host,
				Reporter:  ssr,
				LogLine:   dmp["content"].(string),
				Tags:      tags,
				Timestamp: dmp["timestamp"].(time.Time),
			}
		}

	}()
	return s
}
