package main

import "fmt"

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
	Port int
	Env  string
}

func (ac AppConfig) IsProd() bool {
	return ac.Env == "production"
}

func DefaultConfig() AppConfig {
	return AppConfig{
		Port: 8000,
		Env:  "dev",
	}
}

// <!-- models/users.go -->
// const userPwPepper = "u3lx@T!I8gdKLwsB*q8TsCVxI0LW50rF"
// const hmacSecretKey = "yjqRz4166W6@RvFd#b59yGT6uSIsVJh#"

// <!-- models/services.go -->
// db, err := gorm.Open("postgres", connectionInfo)
// ...
// db.LogMode(true)
