package compass

import (
	"testing"

	"github.com/miekg/dns"
)

func testServer() *Server {
	config := &Config{
		Addr:   "localhost:9999",
		Domain: "n2oapp.com.",
	}
	s, _ := NewServer(config)
	go s.ListenAndServe()
	return s
}

type dnsTestCase struct {
	qname  string
	qtype  uint16
	answer []dns.RR
}

func newA(rr string) *dns.A {
	r, _ := dns.NewRR(rr)
	return r.(*dns.A)
}

func assertA(t *testing.T, got dns.RR, expected dns.RR) {
	a := got.(*dns.A)
	b := expected.(*dns.A)
	if a.A.String() != b.A.String() ||
		a.Header().String() != b.Header().String() {
		t.Fatal("Mismatched record. Got %v, expected %v", got, expected)
	}
}

func assertAnswers(t *testing.T, got []dns.RR, expected []dns.RR) {
	for i, g := range got {
		switch g.(type) {
		case *dns.A:
			assertA(t, g, expected[i])
		}
	}
}

func TestServeBoxhost(t *testing.T) {
	s := testServer()
	defer s.Stop()

	testCases := []dnsTestCase{
		{
			qname:  "test.n2oapp.com.",
			qtype:  dns.TypeA,
			answer: []dns.RR{newA("test.n2oapp.com. 300 A " + data["test"])},
		},
	}

	for _, tc := range testCases {
		c := new(dns.Client)
		m := new(dns.Msg)
		m.SetQuestion(tc.qname, tc.qtype)
		r, _, err := c.Exchange(m, "localhost:9999")
		if err != nil {
			t.Fatal("Failed to get DNS response", err)
		}
		assertAnswers(t, r.Answer, tc.answer)
	}
}
