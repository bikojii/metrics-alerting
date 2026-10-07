package config

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type ServerConfig struct{ Address string }

type AgentConfig struct {
	ServerAddress  string
	ReportInterval int
	PollInterval   int
}

func NormalizeServerAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}
	u, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("invalid server address: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", fmt.Errorf("server address must be an HTTP(S) origin")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func (c *AgentConfig) Validate() error {
	maxSeconds := int64((1<<63 - 1) / time.Second)
	if c.PollInterval <= 0 || c.ReportInterval <= 0 || int64(c.PollInterval) > maxSeconds || int64(c.ReportInterval) > maxSeconds {
		return fmt.Errorf("poll and report intervals must be positive and fit time.Duration")
	}
	address, err := NormalizeServerAddress(c.ServerAddress)
	if err != nil {
		return err
	}
	c.ServerAddress = address
	return nil
}

func LoadServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	fs.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP server address")
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
	if fs.NArg() != 0 {
		log.Fatal("unexpected positional arguments")
	}
	return cfg
}

func LoadAgentConfig() *AgentConfig {
	cfg := &AgentConfig{}
	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	fs.StringVar(&cfg.ServerAddress, "a", "http://localhost:8080", "Address of metrics server")
	fs.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	fs.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
	if fs.NArg() != 0 {
		log.Fatal("unexpected positional arguments")
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	return cfg
}
