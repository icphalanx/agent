package packagekit

import (
	"fmt"
	"github.com/icphalanx/agent/types"
)

type PackageCountMetric struct {
	id        string
	humanName string
	humanDesc string

	packageCount uint

	shouldWarn   bool
	warningLevel uint
	dangerLevel  uint
}

func (pcm PackageCountMetric) Id() string {
	return pcm.id
}

func (PackageCountMetric) MetricType() types.MetricType {
	return types.METRICTYPE_UNCOUNTABLE
}

func (pcm PackageCountMetric) Value() uint {
	return pcm.packageCount
}

func (pcm PackageCountMetric) Status() types.MetricStatus {
	switch {
	case !pcm.shouldWarn:
		return types.METRICSTATUS_NONE
	case pcm.packageCount < pcm.warningLevel:
		return types.METRICSTATUS_HEALTHY
	case pcm.packageCount < pcm.dangerLevel:
		return types.METRICSTATUS_WARNING
	default:
		return types.METRICSTATUS_DANGER
	}
}

func (pcm PackageCountMetric) HumanName() string {
	return pcm.humanName
}

func (pcm PackageCountMetric) HumanDesc() string {
	switch {
	case !pcm.shouldWarn:
		return pcm.humanDesc
	case pcm.packageCount < pcm.warningLevel:
		return fmt.Sprintf(`%s This metric is healthy because the current level is below the configured warning level of %d.`, pcm.humanDesc, pcm.warningLevel)
	case pcm.packageCount < pcm.dangerLevel:
		return fmt.Sprintf(`%s This metric is warning because the current level is between the configured warning level of %d and the danger level of %d.`, pcm.humanDesc, pcm.warningLevel, pcm.dangerLevel)
	default:
		return fmt.Sprintf(`%s This metric is alerting because the current level is above the danger level of %d.`, pcm.humanDesc, pcm.dangerLevel)
	}
}
