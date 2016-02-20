package reporters

import (
	"github.com/icphalanx/agent/types"
	"log"
)

var registry []types.ReporterFactory = []types.ReporterFactory{}

func Register(r types.ReporterFactory) error {
	registry = append(registry, r)
	return nil
}

func GenerateFor(h types.Host) []types.Reporter {
	ret := make([]types.Reporter, 0)
	for _, rf := range registry {
		applicable, err := rf.ApplicableTo(h)
		if err != nil {
			log.Println(rf.Id(), "not applicable to", h.Id(), err)
			continue
		} else if !applicable {
			log.Println(rf.Id(), "not applicable to", h.Id(), "for an unknown reason")
			continue
		}

		log.Println(rf.Id(), "applicable to", h.Id())
		r, err := rf.Create(h)
		if err != nil {
			log.Println(rf.Id(), "error instantiating", err)
			continue
		}
		ret = append(ret, r)
	}
	return ret
}
