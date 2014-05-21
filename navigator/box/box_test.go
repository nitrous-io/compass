package box

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/miekg/dns"
	n "github.com/nitrous-io/compass/navigator"
)

var dbStructure = []string{
	`DROP TABLE IF EXISTS boxes;`,
	`CREATE TABLE boxes (
		id integer NOT NULL,
		name character varying(255) NOT NULL,
		host character varying(255));`,
	`CREATE SEQUENCE boxes_id_seq START WITH 1 INCREMENT BY 1
		NO MINVALUE NO MAXVALUE CACHE 1;`,
	`ALTER SEQUENCE boxes_id_seq OWNED BY boxes.id;`,
	`ALTER TABLE ONLY boxes ALTER COLUMN id SET DEFAULT nextval('boxes_id_seq'::regclass);`,
	`ALTER TABLE ONLY boxes ADD CONSTRAINT boxes_pkey PRIMARY KEY (id);`,
}

var fixtures = []*Box{
	{Name: "daniel-box", Host: "use1.nitrousbox.com"},
	{Name: "md-box", Host: "usw1.nitrousbox.com"},
	{Name: "dqminh-box", Host: "apse1.nitrousbox.com"},
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

func insertBoxhost(db *sql.DB, values []*Box) error {
	stm, err := db.Prepare(`INSERT INTO boxes (name, host) VALUES ($1, $2)
		RETURNING ID`)
	if err != nil {
		return fmt.Errorf("stm error %v", err)
	}
	for _, b := range values {
		err := stm.QueryRow(b.Name, b.Host).Scan(&b.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestFindByName(t *testing.T) {
	n := testNavigator("nitrousbox.com.").(*Navigator)
	host, err := n.FindByName(fixtures[0].containerName())
	if err != nil || host.Name != "daniel-box" || host.Host !=
		"use1.nitrousbox.com" {
		t.Fatal("Failed to find host")
	}
	_, err = n.FindByName("invalid-box")
	if err == nil {
		t.Fatal("should not be able to find invalid box")
	}
}

func TestLookup(t *testing.T) {
	n := testNavigator("nitrousbox.com.").(*Navigator)
	b := fixtures[0]
	r, err := n.Lookup(b.containerName() + ".nitrousbox.com.")
	if err != nil || r[0].(*dns.CNAME).Target != (b.Host+".") {
		t.Fatal("Failed to get correct dns record")
	}
	_, err = n.Lookup("invalid-box.nitrousbox.com.")
	if err == nil {
		t.Fatal("should not be able to look up invalid box")
	}
	_, err = n.Lookup("daniel-box.nitrous.io.")
	if err == nil {
		t.Fatal("should not be able to look up invalid domain")
	}
}
