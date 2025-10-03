// +build integration

package testing

import "os"

// CloudDBConfig returns configuration for the existing Cloud SQL database
func CloudDBConfig() *TestDBConfig {
	// Check if we should use cloud database
	if os.Getenv("USE_CLOUD_DB") == "true" {
		return &TestDBConfig{
			Host:     "162.222.181.26",
			Port:     "5432",
			User:     "postgres",
			Password: "/R|P8JqlSJ[r%cl}",
			DBName:   "postgres",
			SSLMode:  "require",
		}
	}

	// Default to local test database
	return DefaultTestDBConfig()
}

// GetTestDBConfig returns the appropriate database config based on environment
func GetTestDBConfig() *TestDBConfig {
	// Priority: Cloud DB > Environment variables > Local Docker
	if os.Getenv("USE_CLOUD_DB") == "true" {
		return CloudDBConfig()
	}

	return DefaultTestDBConfig()
}
