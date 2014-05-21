package navigator

import "testing"

func TestDomainName(t *testing.T) {
	data := map[string]struct {
		id   int
		name string
	}{
		"test-1":     {1, "test"},
		"test-123-1": {1, "test-123"},
	}
	for expected, s := range data {
		name := DomainName(s.id, s.name)
		if name != expected {
			t.Fatal("failed to get %v, got %v", expected, name)
		}
	}
}

func TestParseDomainName(t *testing.T) {
	data := []struct {
		container string
		name      string
		id        int
	}{
		{"test-box-252", "test-box", 252},
		{"test-box-59-252", "test-box-59", 252},
		{"test-box", "test-box", 0},
	}
	for _, d := range data {
		name, id := ParseDomainName(d.container)
		if name != d.name || id != d.id {
			t.Fatal("failed to parse container %v, got %v %v", d, name, id)
		}
	}
}
