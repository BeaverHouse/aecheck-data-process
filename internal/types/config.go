package types

// Config holds all application configuration
type Config struct {
	Postgres PostgresConfig
	Storage  StorageConfig
	Wiki     WikiConfig
}

// Configuration for the Postgres database
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// StorageConfig holds object storage configuration
type StorageConfig struct {
	UploadURL string
}

// WikiConfig holds wiki scraping configuration
type WikiConfig struct {
	BaseURL   string
	UserAgent string
}
