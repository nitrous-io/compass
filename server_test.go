package compass

import (
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestServeBoxhost(t *testing.T) {
	s, _ := NewServer(nil)
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			t.Fatal("Failed to run DNS server", err)
		}
	}()
	time.Sleep(1 * time.Second)

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion("test.nitrousbox.io.", dns.TypeA)
	r, _, _ := c.Exchange(m, addr)

	a := r.Answer[0].(*dns.A).A
	if a.String() != data["test"] {
		t.Fatal("Unexpected a record for test.nitrous.io", a)
	}
}
