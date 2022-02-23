package config

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"database"`
	Token    Token    `json:"token"`
	Cache    Cache    `json:"cache"`
}

type Server struct {
	Port int `json:"port,omitempty"`
}

type Database struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Port     int    `json:"port,omitempty"`
	Host     string `json:"host,omitempty"`
	Database string `json:"database,omitempty"`
}

type Cache struct {
	Address  string `json:"address,omitempty"`
	Password string `json:"password,omitempty"`
	Database int    `json:"database,omitempty"`
}

type Token struct {
	Secret string `json:"secret,omitempty"`
}
