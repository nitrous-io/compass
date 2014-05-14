package compass

import (
	n "github.com/nitrous-io/compass/navigator"
	_ "github.com/nitrous-io/compass/navigator/dummy"
)

func NewNavigator(provider string, domain string, config map[string]string) (n.Navigator, error) {
	return n.Init(provider, domain, config)
}
