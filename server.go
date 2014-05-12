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
	DatabaseUrl string
}

type Server struct {
	config *Config
	db     *sql.DB
	mux    *dns.ServeMux
	wg     *sync.WaitGroup
}

var (
	addr       = "localhost:6666"
	baseDomain = "nitrousbox.io"
	boxHosts   []*Boxhost
	defaultTTL = uint32(300)

	// initialize some test data now
	// do we can perform things like `dig test.nitrousbox.io`
	data = map[string]string{
		"test":  "10.0.0.1",
		"hello": "10.0.0.2",
		"world": "10.0.0.3",
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
	go s.createServer("tcp", addr)
	go s.createServer("udp", addr)
	s.wg.Wait()
	return nil
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	q := r.Question[0]

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

		w.WriteMsg(m)
	}()

	switch q.Qtype {
	case dns.TypeA:
		var host *Boxhost
		for _, h := range boxHosts {
			domain := h.Name + "." + baseDomain + "."
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
