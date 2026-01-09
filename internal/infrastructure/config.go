package infrastructure

import (
	"database/sql"
	"fmt"
	"net"
	"os"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Config holds shared infrastructure configuration
type Config struct {
	DB     *sql.DB
	Logger *logrus.Logger
}

// NewConfig creates a new infrastructure config with DB connection and logger
func NewConfig() *Config {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Initialize database
	db, err := sql.Open("postgres", getConnectionString())
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	if err := db.Ping(); err != nil {
		logger.WithError(err).Fatal("Failed to ping database")
	}

	logger.Info("Connected to PostgreSQL database")

	return &Config{
		DB:     db,
		Logger: logger,
	}
}

func getConnectionString() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "postgres")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Try to resolve an IPv4 address for the host and include it as hostaddr.
	// This helps avoid "network is unreachable" when the hostname resolves to an IPv6 address
	// but the local machine has no IPv6 connectivity. Keeping `host` preserves SSL hostname verification.
	hostaddr := ""
	ips, err := net.LookupIP(host)
	if err == nil {
		for _, ip := range ips {
			if ip.To4() != nil {
				hostaddr = ip.String()
				break
			}
		}
	}

	if hostaddr != "" {
		return fmt.Sprintf("host=%s hostaddr=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, hostaddr, port, user, password, dbname, sslmode)
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}