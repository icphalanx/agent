package types

import (
	pb "github.com/icphalanx/rpc"
	google_protobuf "google/protobuf"

	"time"
)

func HostToRPC(h Host) (*pb.Host, error) {
	parents := []string{}
	p, err := h.Parent()
	for p != nil {
		parents = append(parents, p.Id())
		p, err = p.Parent()
	}

	hn, err := h.HumanName()

	return &pb.Host{
		Id:        h.Id(),
		HumanName: hn,
		Parents:   parents,
	}, err
}

func ReportersToRPC(rs []Reporter) ([]*pb.Reporter, error) {
	prs := make([]*pb.Reporter, len(rs))
	for n, r := range rs {
		pr, err := ReporterToRPC(r)
		if err != nil {
			return nil, err
		}
		prs[n] = pr
	}
	return prs, nil
}

func ReporterToRPC(r Reporter) (*pb.Reporter, error) {
	pr := new(pb.Reporter)

	if issues, err := r.Issues(); err != nil {
		return nil, err
	} else {
		pr.Issues, err = IssuesToRPC(issues)
		if err != nil {
			return nil, err
		}
	}

	if metrics, err := r.Metrics(); err != nil {
		return nil, err
	} else {
		pr.Metrics, err = MetricsToRPC(metrics)
		if err != nil {
			return nil, err
		}
	}

	// TODO(lukegb): hosts

	return pr, nil
}

func IssuesToRPC(is []Issue) ([]*pb.Issue, error) {
	pis := make([]*pb.Issue, len(is))
	for n, i := range is {
		pis[n] = new(pb.Issue)
		pis[n].Id = "???"
		_ = i
		// TODO(lukegb)
	}
	return pis, nil
}

func MetricsToRPC(ms []Metric) ([]*pb.Metric, error) {
	pms := make([]*pb.Metric, len(ms))
	for n, m := range ms {
		pms[n] = new(pb.Metric)
		pms[n].Id = m.Id()
		pms[n].HumanName = m.HumanName()
		pms[n].HumanDesc = m.HumanDesc()
		pms[n].Type = MetricTypeToRPC(m.MetricType())
		switch m.MetricType() {
		case METRICTYPE_UNCOUNTABLE:
			pms[n].Value = &pb.Metric_IntValue{int64(m.(MetricUncountable).Value())}
		case METRICTYPE_STRINGARRAY:
			pms[n].Value = &pb.Metric_StringArrayValue{&pb.Metric_StringArray{m.(MetricStringArray).Value()}}
		}
	}
	return pms, nil
}

func MetricTypeToRPC(m MetricType) pb.Metric_Type {
	switch m {
	case METRICTYPE_UNCOUNTABLE:
		return pb.Metric_UNCOUNTABLE
	case METRICTYPE_STRINGARRAY:
		return pb.Metric_STRINGARRAY
	}
	return pb.Metric_UNKNOWN
}

func TimeToGoogleTimestamp(t time.Time) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		t.Unix(), int32(t.Nanosecond()),
	}
}
