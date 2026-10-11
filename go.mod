module github.com/cnd-industries/twitch-whip-that-agent

go 1.26.0

require (
	github.com/creack/pty v1.1.24
	github.com/gempir/go-twitch-irc/v4 v4.4.1
	github.com/ilyakaznacheev/cleanenv v1.5.0
	golang.org/x/term v0.47.0
	golang.org/x/time v0.16.0
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/sys v0.49.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

ignore ./web/node_modules
