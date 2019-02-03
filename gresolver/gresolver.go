package gresolver

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
	cachedDns  map[string]*CachedLookup
	cacheMutex *sync.Mutex
}

type CachedLookup struct {
	Name    string
	Expires time.Time
	Address string
}

func New() *Resolver {
	return &Resolver{
		cachedDns:  make(map[string]*CachedLookup),
		cacheMutex: &sync.Mutex{},
	}
}

// Lookup is used by Dial to perform a dns lookup without requiring a dns server
func (r *Resolver) Lookup(address string) (addr string, err error) {

	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()
	if r.cachedDns[address] != nil {
		if time.Now().Before(r.cachedDns[address].Expires) {
			return r.cachedDns[address].Address, nil
		}
		delete(r.cachedDns, address)
	}

	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{Name: address, Qtype: dns.TypeA, Qclass: dns.ClassINET}

	in, err := dns.Exchange(m1, "8.8.8.8:53")
	if err != nil {
		return "", err
	}

	for _, ar := range in.Answer {
		if a, ok := ar.(*dns.A); ok {
			r.cachedDns[address] = &CachedLookup{
				Name:    address,
				Expires: time.Now().Add(time.Second * time.Duration(a.Hdr.Ttl)),
				Address: a.A.String(),
			}
			return a.A.String(), nil
		}
	}
	return "", errors.New("dns lookup failed")
}

// Dial connects to google dns using Lookup and does not require a dns server
func (r *Resolver) Dial(network, address string) (net.Conn, error) {

	ap := strings.Split(address, ":")
	if len(ap) != 2 {
		return nil, fmt.Errorf("expected two address parts, found %d", len(ap))
	}

	ip, err := r.Lookup(ap[0] + ".")
	if err != nil {
		return nil, err
	}

	return net.Dial(network, ip+":"+ap[1])
}
