package boxhost

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	_ "github.com/lib/pq"
	"github.com/miekg/dns"
	nav "github.com/nitrous-io/compass/navigator"
)

const defaultTTL = uint32(300)

var (
	ErrNotFound = errors.New("RecordNotFound")
)

type Boxhost struct {
	Id        int
	Name      string
	IpAddress string
}

type Navigator struct {
	domain string
	db     *sql.DB
}

func (n *Navigator) Lookup(name string) ([]dns.RR, error) {
	if !sameRootDomain(name, n.domain) {
		return []dns.RR{}, fmt.Errorf("cant handle domain %v", name)
	}
	labels := dns.SplitDomainName(name)
	llabel := len(labels)
	if len(labels) < 3 {
		return []dns.RR{}, fmt.Errorf("missing box name %v", name)
	}
	hostName := labels[llabel-3]
	boxHost, err := n.FindByName(hostName)
	if err != nil {
		return []dns.RR{}, fmt.Errorf("Invalid boxname from domain %v", name)
	}
	return []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    defaultTTL,
			},
			A: net.ParseIP(boxHost.IpAddress),
		},
	}, nil
}

func (n *Navigator) FindByName(name string) (*Boxhost, error) {
	_, id := nav.ParseDomainName(name)
	host := new(Boxhost)
	err := n.db.QueryRow(
		"SELECT id, name, ip_address FROM boxhosts WHERE id = $1 LIMIT 1",
		id).Scan(&host.Id, &host.Name, &host.IpAddress)
	if err != nil {
		return nil, err
	}
	return host, nil
}

func (n *Navigator) NewSOA() []dns.RR {
	return []dns.RR{
		&dns.SOA{
			Hdr:     dns.RR_Header{Name: n.domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: defaultTTL},
			Ns:      "ns1." + n.domain,
			Mbox:    "admin." + n.domain,
			Serial:  uint32(time.Now().Truncate(time.Hour).Unix()),
			Refresh: 28800,
			Retry:   7200,
			Expire:  604800,
			Minttl:  defaultTTL,
		},
	}
}

func New(domain string, config map[string]string) (nav.Navigator, error) {
	db, err := sql.Open("postgres", config["database_url"])
	if err != nil {
		return nil, fmt.Errorf("Boxhost: failed to get db connection %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Boxhost: failed to get db connection %v", err)
	}
	return &Navigator{
		domain: domain,
		db:     db,
	}, nil
}

func sameRootDomain(request string, expected string) bool {
	r := dns.SplitDomainName(request)
	lr := len(r)
	e := dns.SplitDomainName(expected)
	le := len(e)
	return r[lr-1] == e[le-1] && r[lr-2] == e[le-2]
}

func init() {
	nav.Navigators["boxhost"] = New
}
