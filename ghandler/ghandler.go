package ghandler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/mlctrez/googledns"
)

func New() *DnsServer {
	d := &DnsServer{
		gdns:       googledns.New(),
		cache:      make(map[string]*CachedQuery),
		cacheMutex: &sync.Mutex{},
	}
	go d.purgeCache()
	return d
}

type CachedQuery struct {
	Expires time.Time
	Answers []dns.RR
}

type DnsServer struct {
	gdns       *googledns.Gdns
	cache      map[string]*CachedQuery
	cacheMutex *sync.Mutex
}

func lowestTtl(ans []dns.RR) time.Duration {
	var lowest uint32
	for _, a := range ans {
		if lowest == 0 || a.Header().Ttl < lowest {
			lowest = a.Header().Ttl
		}
	}
	return time.Duration(lowest)
}

func (d *DnsServer) purgeCache() {
	ticker := time.NewTicker(time.Second * 10)
	for range ticker.C {
		for k, v := range d.cache {
			if v.Expires.Before(time.Now()) {
				d.cacheMutex.Lock()
				delete(d.cache, k)
				d.cacheMutex.Unlock()
			}
		}
	}
}

func (d *DnsServer) cacheAnswer(query string, ans []dns.RR) (result []dns.RR) {
	if len(ans) > 0 {
		d.cacheMutex.Lock()
		defer d.cacheMutex.Unlock()
		loTtl := lowestTtl(ans)
		if loTtl < 10 {
			loTtl = 10
		}
		for _, a := range ans {
			a.Header().Ttl = uint32(loTtl)
		}
		d.cache[query] = &CachedQuery{
			Expires: time.Now().Add(loTtl * time.Second),
			Answers: ans,
		}
		return ans
	}
	return ans
}

func (d *DnsServer) Handler(w dns.ResponseWriter, r *dns.Msg) {

	m := new(dns.Msg)
	m.SetReply(r)
	m.Answer = make([]dns.RR, 0)

	for _, q := range r.Question {
		cr := d.cache[q.String()]
		if cr != nil {
			if cr.Expires.After(time.Now()) {
				m.Answer = append(m.Answer, cr.Answers...)
				continue
			}
			delete(d.cache, q.String())
		}

		ans, err := d.gdns.Query(q)
		if err == nil {
			m.Answer = append(m.Answer, d.cacheAnswer(q.String(), ans)...)
		} else {
			fmt.Println(err)
		}
	}
	err := w.WriteMsg(m)
	if err != nil {
		log.Println("error", err)
	}
}
