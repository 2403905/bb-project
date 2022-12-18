package config

import (
	"os"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	LogLevel       int
	LogPretty      bool
	ApiListener    string
	Postgres       PostgresConfig
	Kafka          KafkaConfig
	ObjectEndpoint string
}

func (c Config) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.ApiListener, v.Required),
		v.Field(&c.LogLevel, v.Min(-1), v.Max(7)),
		v.Field(&c.ObjectEndpoint, v.Required, is.RequestURI),
	)
}

type PostgresConfig struct {
	DSN   string
	Debug bool
}

func (c PostgresConfig) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.DSN, v.Required),
	)
}

type KafkaConfig struct {
	Host    string
	GroupId string
	Topic   string
}

func (c KafkaConfig) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Host, v.Required),
		v.Field(&c.GroupId, v.Required),
		v.Field(&c.Topic, v.Required),
	)
}

// InitConfig
// The .env file is for local running
// For production running use the environment variables
func InitConfig() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn().Err(err).Msg("Can't load the local config file '.env'. \n" +
			"Attempt to load the configuration from the environment variables")
	}
	viper.AutomaticEnv()
	c := new(Config)
	c.LogLevel = viper.GetInt("LOG_LEVEL")
	c.LogPretty = viper.GetBool("LOG_PRETTY")
	c.ApiListener = viper.GetString("API_LISTENER")
	c.Postgres.DSN = viper.GetString("PG_DSN")
	c.Postgres.Debug = viper.GetBool("PG_DEBUG")
	c.Kafka.Host = viper.GetString("KAFKA_HOST")
	c.Kafka.GroupId = viper.GetString("KAFKA_GROUP_ID")
	c.Kafka.Topic = viper.GetString("KAFKA_TOPIC")
	c.ObjectEndpoint = viper.GetString("OBJECT_ENDPOINT")

	if err := c.Validate(); err != nil {
		log.Error().Err(err).Send()
		os.Exit(-1)
	}

	return c
}
