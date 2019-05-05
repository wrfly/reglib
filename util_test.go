package reglib

import "testing"

func TestParseDockerConfig(t *testing.T) {
	c, err := parseDockerConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", c)
}
