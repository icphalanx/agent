package main

import (
	"flag"
	"github.com/icphalanx/agent"
	"log"
)

func main() {
	flag.Parse()

	host, err := agent.Run()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(host)

	reporters, err := host.Reporters()
	if err != nil {
		log.Fatalln(err)
	}

	for _, reporter := range reporters {
		log.Println(reporter.Id())
		m, e := reporter.Metrics()
		log.Println("\t", m, e)
	}
}
