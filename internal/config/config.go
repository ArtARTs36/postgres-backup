package config

import (
	"github.com/artarts36/postgres-backup/internal/storage"
	"github.com/caarlos0/env/v11"
)

type PostgresConfig struct {
	Host     string   `env:"HOST,required"`
	Port     int      `env:"PORT,required"`
	User     string   `env:"USER,required"`
	Password string   `env:"PASSWORD_FILE,required,file,notEmpty,unset"` //nolint:gosec // false-positive: no json
	Database []string `env:"DATABASE,required,notEmpty"`
}

type S3Config struct {
	Endpoint  string `env:"ENDPOINT,required"`
	AccessKey string `env:"ACCESS_KEY_FILE,required,file,notEmpty,unset"` //nolint:gosec // false-positive: no json
	SecretKey string `env:"SECRET_KEY_FILE,required,file,notEmpty,unset"` //nolint:gosec // false-positive: no json
	UseSSL    bool   `env:"USE_SSL" envDefault:"true"`
	Bucket    string `env:"BUCKET,required"`
}

const (
	StorageTypeS3 = "s3"
	StorageTypeFS = "fs"
)

type Config struct {
	Postgres      PostgresConfig   `envPrefix:"PG_"`
	S3            storage.S3Config `envPrefix:"STORAGE_S3_"`
	TempDir       string           `env:"TEMP_DIR" envDefault:"/tmp"`
	MaxBackups    int              `env:"MAX_BACKUPS" envDefault:"10"`
	MetricsServer string           `env:"METRICS_SERVER"`

	StorageType string `env:"STORAGE_TYPE" envDefault:"s3"`
	FSRoot      string `env:"FS_ROOT"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
