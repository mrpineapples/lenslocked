package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (pc PostgresConfig) Dialect() string {
	return "postgres"
}

func (pc PostgresConfig) ConnectionInfo() string {
	if pc.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", pc.Host, pc.Port, pc.User, pc.Name)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pc.Host, pc.Port, pc.User, pc.Password, pc.Name)
}

func DefaultPosgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "Michael",
		Password: "not-necessary",
		Name:     "lenslocked_dev",
	}
}

type AppConfig struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
	Mailgun  MailgunConfig  `json:"mailgun"`
	Dropbox  OAuthConfig    `json:"dropbox"`
}

func (ac AppConfig) IsProd() bool {
	return ac.Env == "production"
}

func DefaultConfig() AppConfig {
	return AppConfig{
		Port:     8000,
		Env:      "dev",
		Pepper:   "u3lx@T!I8gdKLwsB*q8TsCVxI0LW50rF",
		HMACKey:  "yjqRz4166W6@RvFd#b59yGT6uSIsVJh#",
		Database: DefaultPosgresConfig(),
	}
}

type MailgunConfig struct {
	APIKey       string `json:"api_key"`
	PublicAPIKey string `json:"public_api_key"`
	Domain       string `json:"domain"`
}

type OAuthConfig struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	AuthURL  string `json:"auth_url"`
	TokenURL string `json:"token_url"`
}

func LoadConfig(configReq bool) AppConfig {
	file, err := os.Open(".config.json")
	if err != nil {
		if configReq {
			panic(err)
		}
		fmt.Println("Using the default config...")
		return DefaultConfig()
	}

	var c AppConfig
	dec := json.NewDecoder(file)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully loaded .config.json")
	return c
}
