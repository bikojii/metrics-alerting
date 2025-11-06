package config

import (
	"flag"
	"log"
	"os"
)

type ServerConfig struct {
	Address string //адрес http сервера
}

type AgentConfig struct {
	ServerAddress  string // адрес сервера, куда агент отправляет метрики
	ReportInterval int    // интервал отправки метрик
	PollInterval   int    // интервал опроса метрик
}

// для сервера
func LoadServerConfig() *ServerConfig {
	cfg := &ServerConfig{}

	fs := flag.NewFlagSet("server", flag.ExitOnError)
	fs.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP server address")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Ошибка парсинга флагов сервера: %v", err)
	}

	return cfg
}

func LoadAgentConfig() *AgentConfig {
	cfg := &AgentConfig{}

	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	fs.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Address of metrics server")
	fs.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	fs.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Ошибка парсинга флагов агента: %v", err)
	}

	return cfg
}
