package main

import (
	"flag"
	"github.com/icphalanx/agent"
	"log"
)

var (
	Upstream = flag.String("upstream", "127.0.0.1:13890", "upstream location")

	CertLocation    = flag.String("certLocation", "phagent.crt", "location to store certificate")
	PrivKeyLocation = flag.String("privKeyLocation", "phagent.key", "location to store private key")

	CALocation = flag.String("caLocation", "phalanx.crt", "location of CA")
)

func main() {
	flag.Parse()

	reporter, err := agent.NewRPCAgent(*Upstream, *CALocation, *CertLocation, *PrivKeyLocation)
	if err != nil {
		log.Fatalln(err)
	}

	log.Fatalln(reporter.Run())
}
