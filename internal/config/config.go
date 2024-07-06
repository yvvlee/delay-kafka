package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// NewConfig init config
func NewConfig() *Config {
	return &Config{
		Redis: &Redis{
			Addr:         getEnv("DELAY_KAFKA_REDIS_ADDR", ""),
			Password:     getEnv("DELAY_KAFKA_REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("DELAY_KAFKA_REDIS_DB", 0),
			ReadTimeout:  getEnvAsDuration("DELAY_KAFKA_REDIS_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getEnvAsDuration("DELAY_KAFKA_REDIS_WRITE_TIMEOUT", 5*time.Second),
			PoolSize:     getEnvAsInt("DELAY_KAFKA_REDIS_POOL_SIZE", 10),
		},
		Kafka: &kafka{
			Brokers:       strings.Split(getEnv("DELAY_KAFKA_BROKERS", ""), ","),
			Topic:         getEnv("DELAY_KAFKA_TOPIC", ""),
			ConsumerGroup: getEnv("DELAY_KAFKA_CONSUMER_GROUP", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return fallback
}

type Config struct {
	Redis *Redis
	Kafka *kafka
}

type Redis struct {
	Addr         string
	Password     string
	DB           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
}

type kafka struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
}
