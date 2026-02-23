// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package config loads and validates the application configuration from a YAML file.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config is the root configuration struct.
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Kafka        KafkaConfig        `mapstructure:"kafka"`
	Engine       EngineConfig       `mapstructure:"engine"`
	Feature      FeatureConfig      `mapstructure:"feature"`
	FeatureStore FeatureStoreConfig `mapstructure:"feature_store"`
	Log          LogConfig          `mapstructure:"log"`
}

// ServerConfig holds HTTP/gRPC server settings.
type ServerConfig struct {
	Addr         string        `mapstructure:"addr"`
	GRPCAddr     string        `mapstructure:"grpc_addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	DrainTimeout time.Duration `mapstructure:"drain_timeout"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addrs    []string `mapstructure:"addrs"`
	Password string   `mapstructure:"password"`
	DB       int      `mapstructure:"db"`
	PoolSize int      `mapstructure:"pool_size"`
}

// KafkaConfig holds Kafka producer settings for the audit writer.
type KafkaConfig struct {
	Brokers    []string `mapstructure:"brokers"`
	AuditTopic string   `mapstructure:"audit_topic"`
}

// EngineConfig holds top-level engine settings.
type EngineConfig struct {
	// PolicyDir is the directory containing PolicySet YAML files.
	PolicyDir    string        `mapstructure:"policy_dir"`
	ReloadInterval time.Duration `mapstructure:"reload_interval"`
	// Fallback is the global fallback decision when a scene is not found.
	Fallback     string        `mapstructure:"fallback"`
}

// FeatureConfig holds settings for the feature service.
type FeatureConfig struct {
	TotalTimeout    time.Duration `mapstructure:"total_timeout"`
	RedisTimeout    time.Duration `mapstructure:"redis_timeout"`
	ExternalTimeout time.Duration `mapstructure:"external_timeout"`
}

// FeatureStoreConfig holds settings for the standalone Feature Store service.
type FeatureStoreConfig struct {
	// Enabled controls whether the decision engine connects to an external
	// Feature Store. When false, only in-process fetchers are used.
	Enabled bool `mapstructure:"enabled"`
	// Addr is the Feature Store gRPC endpoint (e.g. "feature-store:9100").
	// Also used as the listen address when running cmd/featurestore.
	Addr string `mapstructure:"addr"`
	// DialTimeout is the maximum time to wait for the initial connection.
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
	// RequestTimeout is the per-RPC deadline for feature fetch calls.
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	// Groups lists the feature groups to fetch from the store.
	// Each group maps to one FeatureStoreFetcher registered in feature.Service.
	Groups []FeatureStoreGroupConfig `mapstructure:"groups"`
}

// FeatureStoreGroupConfig configures one feature group fetched from the store.
type FeatureStoreGroupConfig struct {
	// Name must match a registered FeatureGroup on the server side.
	Name string `mapstructure:"name"`
	// Timeout overrides the global RequestTimeout for this group.
	Timeout time.Duration `mapstructure:"timeout"`
}

// LogConfig holds logger settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, console
}

// Load reads and validates the configuration at the given path.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.grpc_addr", ":9090")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.drain_timeout", "30s")
	v.SetDefault("redis.pool_size", 50)
	v.SetDefault("kafka.audit_topic", "risk.decisions")
	v.SetDefault("engine.policy_dir", "configs/policies")
	v.SetDefault("engine.reload_interval", "30s")
	v.SetDefault("engine.fallback", "PASS")
	v.SetDefault("feature.total_timeout", "50ms")
	v.SetDefault("feature.redis_timeout", "10ms")
	v.SetDefault("feature.external_timeout", "30ms")
	v.SetDefault("feature_store.enabled", false)
	v.SetDefault("feature_store.addr", ":9100")
	v.SetDefault("feature_store.dial_timeout", "5s")
	v.SetDefault("feature_store.request_timeout", "20ms")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}
