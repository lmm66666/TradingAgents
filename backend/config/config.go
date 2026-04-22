package config

type Config struct {
	DB *DB
}

type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}
