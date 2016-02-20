package types

type Host interface {
	Id() string

	IsLocal() bool

	HumanName() (string, error)

	Parent() (Host, error)

	Reporters() ([]Reporter, error)
}

type Issue interface {
}

type Metric interface {
	Id() string

	HumanName() string
	HumanDesc() string

	MetricType() MetricType

	Status() MetricStatus
}

type MetricUncountable interface {
	Metric

	Value() int
}

type MetricStringArray interface {
	Metric

	Value() []string
}

type MetricType uint

const (
	METRICTYPE_UNCOUNTABLE = iota
	METRICTYPE_STRINGARRAY
)

type MetricStatus uint

const (
	METRICSTATUS_NONE = iota
	METRICSTATUS_HEALTHY
	METRICSTATUS_WARNING
	METRICSTATUS_DANGER
)

type ReporterFactory interface {
	Id() string

	ApplicableTo(Host) (bool, error)

	Create(Host) (Reporter, error)
}

type Reporter interface {
	Id() string

	// returns a list of "issues" affecting this host
	Issues() ([]Issue, error)

	// returns a list of "metrics" affecting this host
	Metrics() ([]Metric, error)

	// returns a list of subhosts under this host
	Hosts() ([]Host, error)
}
