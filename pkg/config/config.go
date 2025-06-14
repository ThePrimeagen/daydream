package config

const SERVER_SOCKET = "/tmp/opencode-server"

type CLIConfig struct {
	Debug bool
}

var CLIConfigInstance = CLIConfig{
	Debug: false,
}

