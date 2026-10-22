package handlers

import "testing"

func TestParseLimit(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"", defaultFeedLimit},
		{"abc", defaultFeedLimit},
		{"0", defaultFeedLimit},
		{"-5", defaultFeedLimit},
		{"1", 1},
		{"10", 10},
		{"50", maxFeedLimit},
		{"51", maxFeedLimit},
		{"1000", maxFeedLimit},
	}

	for _, c := range cases {
		if got := parseLimit(c.raw); got != c.want {
			t.Errorf("parseLimit(%q) = %d, want %d", c.raw, got, c.want)
		}
	}
}
