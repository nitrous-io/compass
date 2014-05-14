package compass

import (
	"database/sql"
	"log"
	"net"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/miekg/dns"
)

type Boxhost struct {
	Name string
	Host string
}

type Config struct {
	Addr        string `json:"address"`      // host:port
	DatabaseUrl string `json:"database_url"` // postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full
	Domain      string `json:"domain"`       // n2oapp.com.
}

type Server struct {
	config *Config
	db     *sql.DB
	mux    *dns.ServeMux
	wg     *sync.WaitGroup
}

var (
	boxHosts   []*Boxhost
	defaultTTL = uint32(300)

	// initialize some test data now
	// do we can perform things like `dig test.nitrousbox.io`
	data = map[string]string{
		"ns1":   "128.199.236.54",
		"ns2":   "128.199.236.54",
		"test":  "173.194.117.8",
		"www":   "173.194.117.8",
		"hello": "173.194.117.8",
		"world": "173.194.117.8",
	}
)

func init() {
	for name, ip := range data {
		boxHosts = append(boxHosts, &Boxhost{Name: name, Host: ip})
	}
}

func (s *Server) ListenAndServe() error {
	s.mux = dns.NewServeMux()
	s.mux.Handle(".", s)

	s.wg.Add(2)
	go s.createServer("tcp", s.config.Addr)
	go s.createServer("udp", s.config.Addr)
	s.wg.Wait()
	return nil
}

// TODO: ideally we want this to
// - clean up the goroutine
// - exit the ListenAndServe routine
func (s *Server) Stop() {
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	q := r.Question[0]
	log.Println("Question", q, q.Qtype, q.Name)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = make([]dns.RR, 0, 10)

	defer func() {
		if len(m.Answer) > 1 {
			for _, r := range m.Answer {
				r.Header().Ttl = uint32(defaultTTL)
			}
		}

		log.Println(q.Qtype, q.Name, m.Answer)

		w.WriteMsg(m)
	}()
	if q.Name == s.config.Domain && q.Qtype == dns.TypeSOA {
		log.Println("got soa", q.Name, s.SOA())
		m.Answer = []dns.RR{s.SOA()}
		return
	}

	switch q.Qtype {
	case dns.TypeA:
		var host *Boxhost
		for _, h := range boxHosts {
			domain := h.Name + "." + s.config.Domain
			if domain == q.Name {
				host = h
			}
		}
		if host != nil {
			records := []dns.RR{}
			addr := net.ParseIP(host.Host)
			record := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    defaultTTL,
				},
				A: addr,
			}
			records = append(records, record)
			m.Answer = append(m.Answer, records...)
		}
	}
	if len(m.Answer) == 0 {
		m.Ns = []dns.RR{s.SOA()}
		m.Ns[0].Header().Ttl = defaultTTL
	}
}

func (s *Server) NS() dns.RR {
	return &dns.NS{
		Hdr: dns.RR_Header{Name: s.config.Domain, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: defaultTTL},
		Ns:  "ns1." + s.config.Domain,
	}
}

// SOA returns a SOA record for this SkyDNS instance.
func (s *Server) SOA() dns.RR {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: s.config.Domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: defaultTTL},
		Ns:      "ns1." + s.config.Domain,
		Mbox:    "admin." + s.config.Domain,
		Serial:  uint32(time.Now().Truncate(time.Hour).Unix()),
		Refresh: 28800,
		Retry:   7200,
		Expire:  604800,
		Minttl:  defaultTTL,
	}
}

func (s *Server) createServer(net, addr string) {
	defer s.wg.Done()

	server := &dns.Server{
		Net:         net,
		Addr:        addr,
		Handler:     s.mux,
		ReadTimeout: 1 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func NewServer(config *Config) (*Server, error) {
	server := &Server{
		config: config,
		wg:     new(sync.WaitGroup),
	}
	return server, nil
}
