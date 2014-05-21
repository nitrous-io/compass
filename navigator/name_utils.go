package navigator

import (
	"fmt"
	"strconv"
	"strings"
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
