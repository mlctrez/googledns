package googledns

import (
	"fmt"

	"github.com/miekg/dns"
)

const GOOGLE_DNS_API = "https://dns.google.com/resolve"

type Response struct {
	Status   uint32
	TC       bool
	RD       bool
	RA       bool
	AD       bool
	CD       bool
	Question []Question
	Answer   []Answer
}

type Question struct {
	Name string
	Type uint16
}

type Answer struct {
	Name string
	Type uint16
	TTL  uint32
	Data string
}

func (a Answer) ToRR() (rr dns.RR, ts string, err error) {

	ts = dns.TypeToString[a.Type]
	if ts == "" {
		return rr, ts, fmt.Errorf("unhandled %d", a.Type)
	}

	rrs := fmt.Sprintf("%s %d %s %s %s", a.Name, a.TTL, "IN", ts, a.Data)
	rr, err = dns.NewRR(rrs)
	return
}
