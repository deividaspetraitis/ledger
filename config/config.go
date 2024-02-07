package config

import (
	"strings"

	"github.com/deividaspetraitis/ledger/http"

	"github.com/deividaspetraitis/go/database"
	"github.com/deividaspetraitis/go/errors"

	"github.com/spf13/viper"
)

// ErrConfigNotFound represents an error returned when configuration file was not found.
var ErrConfigNotFound = errors.New("config file were not found")

// Config represents application configuration.
type Config struct {
	HTTP     *http.Config     `mapstructure:"http"` // HTTP server config.
	Database *database.Config `mapstructure:"db"`   // Database instance config.
}

// New accepts constructs a new Config by reading env configuration file.
func New(path string) (*Config, error) {
	parser := viper.NewWithOptions(
		viper.KeyDelimiter("_"), // DATABASE_HOST instead of DATABASE.HOST in config file
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
	)

	// Set config file
	parser.SetConfigFile(path)

	// Set config type to env files
	parser.SetConfigType("env")

	// Check and load environment variables
	parser.AutomaticEnv()

	// Read configuration values
	if err := parser.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.Wrapf(err, "unable to read config reading config: %s", path)
		} else {
			return nil, errors.Wrapf(err, "failed reading config: %s", path)
		}
	}

	// Populate configuration
	var cfg Config
	if err := parser.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal config: %s", path)
	}

	return &cfg, nil
}
