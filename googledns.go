package googledns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"github.com/mlctrez/googledns/gresolver"
)

// Gdns is the api for googledns
type Gdns struct {
	transp *http.Transport
}

// New constructs a new Gdns
func New() *Gdns {
	grnew := gresolver.New()
	return &Gdns{
		transp: &http.Transport{
			Dial:            grnew.Dial,
			IdleConnTimeout: time.Duration(60 * time.Second),
			MaxIdleConns:    10,
		},
	}
}

// Query calls the google api for a dns query
func (g *Gdns) Query(q dns.Question) (answers []dns.RR, err error) {

	query := fmt.Sprintf("%s?name=%s&type=%d", GOOGLE_DNS_API, q.Name, q.Qtype)

	var resp *http.Response
	var req *http.Request

	req, err = http.NewRequest("GET", query, &bytes.Buffer{})
	if err != nil {
		return
	}

	resp, err = g.transp.RoundTrip(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	gr := &Response{}
	err = json.NewDecoder(resp.Body).Decode(gr)
	if err != nil {
		return
	}

	answers = make([]dns.RR, 0)
	var ts = ""
	var rr dns.RR
	for _, ga := range gr.Answer {
		if rr, ts, err = ga.ToRR(); err == nil {
			answers = append(answers, rr)
		} else {
			return
		}
	}

	if ts == "" {
		ts = dns.TypeToString[q.Qtype]
	}

	return
}
