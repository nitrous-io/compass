package box

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/miekg/dns"
	nav "github.com/nitrous-io/compass/navigator"
)

const defaultTTL = uint32(300)

var (
	ErrNotFound = errors.New("RecordNotFound")
)

type Box struct {
	Id   int
	Name string
	Host string
}

func (b *Box) containerName() string {
	return fmt.Sprintf("%s-%d", b.Name, b.Id)
}

type Navigator struct {
	domain string
	db     *sql.DB
}

func (n *Navigator) Lookup(name string) ([]dns.RR, error) {
	if !nav.SameRootDomain(name, n.domain) {
		return []dns.RR{}, fmt.Errorf("cant handle domain %v", name)
	}
	labels := dns.SplitDomainName(name)
	llabel := len(labels)
	if len(labels) < 3 {
		return []dns.RR{}, fmt.Errorf("missing box name %v", name)
	}
	hostName := labels[llabel-3]
	box, err := n.FindByName(hostName)
	if err != nil {
		return []dns.RR{}, fmt.Errorf("Invalid boxname from domain %v", name)
	}
	return []dns.RR{
		&dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    defaultTTL,
			},
			Target: box.Host + ".", // needs to end in .
		},
	}, nil
}

func (n *Navigator) FindByName(container string) (*Box, error) {
	_, id := nav.ParseDomainName(container)
	if id <= 0 {
		return nil, ErrNotFound
	}
	box := new(Box)
	err := n.db.QueryRow(
		"SELECT id, name, host FROM boxes WHERE id = $1 LIMIT 1",
		id).Scan(&box.Id, &box.Name, &box.Host)
	if err != nil {
		return nil, err
	}
	return box, nil
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
		return nil, fmt.Errorf("Box: failed to get db connection %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Box: failed to ping db %v", err)
	}
	return &Navigator{
		domain: domain,
		db:     db,
	}, nil
}

func init() {
	nav.Navigators["box"] = New
}
