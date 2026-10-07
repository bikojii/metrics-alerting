package config

import "testing"

func TestNormalizeServerAddress(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"localhost:8080", "http://localhost:8080"},
		{"https://example.com/", "https://example.com"},
		{"http://[::1]:8080", "http://[::1]:8080"},
	} {
		got, err := NormalizeServerAddress(tc.input)
		if err != nil || got != tc.want {
			t.Errorf("%q: got %q, %v", tc.input, got, err)
		}
	}
	for _, input := range []string{"", "ftp://example.com", "http://", "http://host/path", "http://host?q=x", "http://user:pass@host"} {
		if _, err := NormalizeServerAddress(input); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
}

func TestValidateIntervals(t *testing.T) {
	for _, tc := range []struct{ poll, report int }{{0, 1}, {-1, 1}, {1, 0}, {1, -1}} {
		cfg := AgentConfig{ServerAddress: "localhost:8080", PollInterval: tc.poll, ReportInterval: tc.report}
		if cfg.Validate() == nil {
			t.Errorf("accepted intervals %+v", tc)
		}
	}
	cfg := AgentConfig{ServerAddress: "localhost:8080", PollInterval: 2, ReportInterval: 10}
	if err := cfg.Validate(); err != nil || cfg.ServerAddress != "http://localhost:8080" {
		t.Fatalf("config = %+v, error = %v", cfg, err)
	}
}
