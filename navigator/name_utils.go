package navigator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

func DomainName(id int, name string) string {
	return fmt.Sprintf("%s-%d", name, id)
}

// from the name boxName-boxId, return boxName and boxId
func ParseDomainName(domain string) (string, int) {
	lastIndex := strings.LastIndex(domain, "-")
	if lastIndex == -1 {
		return domain, 0
	}
	name := domain[0:lastIndex]
	id, err := strconv.Atoi(domain[lastIndex+1 : len(domain)])
	if err != nil {
		return domain, 0
	}
	return name, id
}

func SameRootDomain(request string, expected string) bool {
	r := dns.SplitDomainName(request)
	lr := len(r)
	e := dns.SplitDomainName(expected)
	le := len(e)
	return r[lr-1] == e[le-1] && r[lr-2] == e[le-2]
}
