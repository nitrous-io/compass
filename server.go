package compass

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/miekg/dns"
)

type Config struct {
	DatabaseUrl string
}

type Server struct {
	config *Config
	db     *sql.DB
	mux    *dns.ServeMux
}

func (s *Server) ListenAndServe() error {
	s.mux = dns.NewServeMux()
	s.mux.Handle(".", s)
	return nil
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
}

func NewServer(config *Config) (*Server, error) {
	db, err := sql.Open("postgres", config.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to postgresl %v", err)
	}
	server := &Server{
		db:     db,
		config: config,
	}
	return server, nil
}
