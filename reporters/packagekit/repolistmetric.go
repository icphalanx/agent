package packagekit

import (
	"github.com/icphalanx/agent/types"
)

type RepoListMetric struct {
	repoList []string
}

func (RepoListMetric) Id() string {
	return "repolist"
}

func (RepoListMetric) MetricType() types.MetricType {
	return types.METRICTYPE_STRINGARRAY
}

func (rlm RepoListMetric) Value() []string {
	return rlm.repoList
}

func (rlm RepoListMetric) Status() types.MetricStatus {
	return types.METRICSTATUS_NONE
}

func (RepoListMetric) HumanName() string {
	return "Package repository list"
}

func (RepoListMetric) HumanDesc() string {
	return "A list of package repositories currently enabled on this host"
}
