package main

import (
	"flag"
	"log"
	"os"

	"github.com/miekg/dns"
	"github.com/mlctrez/googledns/ghandler"
)

func main() {
	listen := flag.String("listen", ":53", "listen address")
	flag.Parse()
	log.SetOutput(os.Stdout)

	dnsServer := ghandler.New()

	server := &dns.Server{Addr: *listen, Net: "udp"}
	dns.HandleFunc(".", dnsServer.Handler)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
