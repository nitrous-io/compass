package boxhost

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/miekg/dns"
	n "github.com/nitrous-io/compass/navigator"
)

var dbStructure = []string{
	`DROP TABLE IF EXISTS boxhosts;`,
	`CREATE TABLE boxhosts (
		id integer NOT NULL,
		name character varying(255) NOT NULL,
		ip_address character varying(255) NOT NULL);`,
	`CREATE SEQUENCE boxhosts_id_seq START WITH 1 INCREMENT BY 1
		NO MINVALUE NO MAXVALUE CACHE 1;`,
	`ALTER SEQUENCE boxhosts_id_seq OWNED BY boxhosts.id;`,
	`ALTER TABLE ONLY boxhosts ALTER COLUMN id SET DEFAULT nextval('boxhosts_id_seq'::regclass);`,
	`ALTER TABLE ONLY boxhosts ADD CONSTRAINT boxhosts_pkey PRIMARY KEY (id);`,
}

var fixtures = []Boxhost{
	{Name: "daniel-box", IpAddress: "10.0.0.1"},
	{Name: "md-box", IpAddress: "10.0.0.2"},
	{Name: "dqminh-box", IpAddress: "10.0.0.3"},
}

func testNavigator(domain string) n.Navigator {
	config := map[string]string{
		"database_url": "host=/var/run/postgresql dbname=testdb sslmode=disable",
	}
	b, err := New(domain, config)
	if err != nil {
		log.Println("failed to create boxhost", err)
	}
	for _, s := range dbStructure {
		_, err = b.(*Navigator).db.Exec(s)
		if err != nil {
			log.Println("failed to set db structure", err, s)
		}
	}
	if err := insertBoxhost(b.(*Navigator).db, fixtures); err != nil {
		log.Println("failed to set db fixtures", err)
	}
	return b
}

func insertBoxhost(db *sql.DB, values []Boxhost) error {
	stm, err := db.Prepare(`INSERT INTO boxhosts (name, ip_address) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("stm error %v", err)
	}
	for _, b := range values {
		_, err := stm.Exec(b.Name, b.IpAddress)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestFindByName(t *testing.T) {
	n := testNavigator("example.com.").(*Navigator)
	host, err := n.FindByName("daniel-box")
	if err != nil || host.Name != "daniel-box" || host.IpAddress != "10.0.0.1" {
		t.Fatal("Failed to find host")
	}
	_, err = n.FindByName("invalid-box")
	if err == nil {
		t.Fatal("should not be able to find invalid box")
	}
}

func TestLookup(t *testing.T) {
	n := testNavigator("example.com.").(*Navigator)
	r, err := n.Lookup("daniel-box.example.com.")
	log.Println(r, err)
	if err != nil || r[0].(*dns.A).A.String() != "10.0.0.1" {
		t.Fatal("Failed to get correct dns record")
	}
	_, err = n.Lookup("invalid-box.example.com.")
	if err == nil {
		t.Fatal("should not be able to look up invalid box")
	}
	_, err = n.Lookup("daniel-box.nitrous.io.")
	if err == nil {
		t.Fatal("should not be able to look up invalid domain")
	}
}
