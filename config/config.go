package config

import (
	"event-tracking-service/pkg/common"
	"log"
	"time"

	"github.com/joho/godotenv"
)

var DefaultSensitiveFields = []string{"password", "token", "secret", "credit_card"}

type DBConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	DBName      string
	SSLMode     string
	MaxIdle     int
	MaxOpen     int
	MaxLife     int
	MaxIdleTime int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	TTL      time.Duration
}

type ServerConfig struct {
	Env  string
	Port string
}

type SchedulerConfig struct {
	ProcessInterval time.Duration
	ProcessTimeout  time.Duration
}

type LogConfig struct {
	Level           string
	Format          string
	ServiceName     string
	Channel         string // console, signoz
	OTLPEndpoint    string
	OTLPToken       string
	EnableGRPC      bool
	EnableAsync     bool
	AsyncBufferSize int
	BatchSize       int
	BatchTimeout    int
}

type TracingConfig struct {
	Enable      bool
	ServiceName string
	Endpoint    string
	UseGRPC     bool
	SampleRatio float64
}

type SentryConfig struct {
	Enable           bool
	DSN              string
	SampleRate       float64
	TracesSampleRate float64
	Debug            bool
}

type EventBufferConfig struct {
	QueueKey        string
	DeadLetterKey   string
	BatchSize       int
	MaxRetries      int
	SensitiveFields []string // Fields to strip from properties before storing
}

type Config struct {
	DB          DBConfig
	Redis       RedisConfig
	Server      ServerConfig
	Scheduler   SchedulerConfig
	EventBuffer EventBufferConfig
	Log         LogConfig
	Tracing     TracingConfig
	Sentry      SentryConfig
}

// ObservabilityConfig returns a unified observability configuration
type ObservabilityConfig struct {
	ServiceName            string
	Environment            string
	LogLevel               string
	LogFormat              string
	LogChannel             string
	OTLPEndpoint           string
	OTLPToken              string
	EnableGRPC             bool
	EnableAsync            bool
	AsyncBufferSize        int
	BatchSize              int
	BatchTimeout           int
	TracingEnable          bool
	TracingEndpoint        string
	TracingUseGRPC         bool
	TracingSampleRatio     float64
	SentryEnable           bool
	SentryDSN              string
	SentrySampleRate       float64
	SentryTracesSampleRate float64
	SentryDebug            bool
}

// GetObservabilityConfig returns unified observability config from existing config
func (c *Config) GetObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		ServiceName:            c.Log.ServiceName,
		Environment:            c.Server.Env,
		LogLevel:               c.Log.Level,
		LogFormat:              c.Log.Format,
		LogChannel:             c.Log.Channel,
		OTLPEndpoint:           c.Log.OTLPEndpoint,
		OTLPToken:              c.Log.OTLPToken,
		EnableGRPC:             c.Log.EnableGRPC,
		EnableAsync:            c.Log.EnableAsync,
		AsyncBufferSize:        c.Log.AsyncBufferSize,
		BatchSize:              c.Log.BatchSize,
		BatchTimeout:           c.Log.BatchTimeout,
		TracingEnable:          c.Tracing.Enable,
		TracingEndpoint:        c.Tracing.Endpoint,
		TracingUseGRPC:         c.Tracing.UseGRPC,
		TracingSampleRatio:     c.Tracing.SampleRatio,
		SentryEnable:           c.Sentry.Enable,
		SentryDSN:              c.Sentry.DSN,
		SentrySampleRate:       c.Sentry.SampleRate,
		SentryTracesSampleRate: c.Sentry.TracesSampleRate,
		SentryDebug:            c.Sentry.Debug,
	}
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		DB: DBConfig{
			Host:        common.GetEnv("DB_HOST", "localhost"),
			Port:        common.GetEnvInt("DB_PORT", 5432),
			Username:    common.GetEnv("DB_USERNAME", ""),
			Password:    common.GetEnv("DB_PASSWORD", ""),
			DBName:      common.GetEnv("DB_DATABASE", ""),
			SSLMode:     common.GetEnv("DB_SSL_MODE", "disable"),
			MaxIdle:     common.GetEnvInt("DB_MAX_IDLE_CONNS", 20),
			MaxOpen:     common.GetEnvInt("DB_MAX_OPEN_CONNS", 100),
			MaxLife:     common.GetEnvInt("DB_CONN_MAX_LIFETIME", 30),
			MaxIdleTime: common.GetEnvInt("DB_CONN_MAX_IDLE_TIME", 10),
		},
		Redis: RedisConfig{
			Host:     common.GetEnv("REDIS_HOST", "localhost"),
			Port:     common.GetEnv("REDIS_PORT", "6379"),
			Password: common.GetEnv("REDIS_PASSWORD", ""),
			DB:       0,
			TTL:      24 * time.Hour,
		},
		Server: ServerConfig{
			Env:  common.GetEnv("APP_ENV", "develop"),
			Port: common.GetEnv("APP_PORT", "8080"),
		},
		Scheduler: SchedulerConfig{
			ProcessInterval: time.Duration(common.GetEnvInt("SCHEDULER_PROCESS_INTERVAL_SECONDS", 60)) * time.Second,
			ProcessTimeout:  time.Duration(common.GetEnvInt("SCHEDULER_PROCESS_TIMEOUT_SECONDS", 60)) * time.Second,
		},
		EventBuffer: EventBufferConfig{
			QueueKey:        common.GetEnv("EVENT_BUFFER_QUEUE_KEY", "event_tracking:events"),
			DeadLetterKey:   common.GetEnv("EVENT_BUFFER_DEAD_LETTER_KEY", "event_tracking:dead_letter"),
			BatchSize:       common.GetEnvInt("EVENT_BUFFER_BATCH_SIZE", 1500),
			MaxRetries:      common.GetEnvInt("EVENT_BUFFER_MAX_RETRIES", 3),
			SensitiveFields: DefaultSensitiveFields,
		},
		Log: LogConfig{
			Level:           common.GetEnv("LOG_LEVEL", "info"),
			Format:          common.GetEnv("LOG_FORMAT", "json"),
			ServiceName:     common.GetEnv("LOG_SERVICE_NAME", "event-tracking"),
			Channel:         common.GetEnv("LOG_CHANNEL", "console"),
			OTLPEndpoint:    common.GetEnv("LOG_OTLP_ENDPOINT", "localhost:4318"),
			OTLPToken:       common.GetEnv("LOG_OTLP_TOKEN", ""),
			EnableGRPC:      common.GetEnvBool("LOG_ENABLE_GRPC", false),
			EnableAsync:     common.GetEnvBool("LOG_ENABLE_ASYNC", true),
			AsyncBufferSize: common.GetEnvInt("LOG_ASYNC_BUFFER_SIZE", 2048),
			BatchSize:       common.GetEnvInt("LOG_BATCH_SIZE", 100),
			BatchTimeout:    common.GetEnvInt("LOG_BATCH_TIMEOUT", 10),
		},
		Tracing: TracingConfig{
			Enable:      common.GetEnvBool("TRACING_ENABLE", false),
			ServiceName: common.GetEnv("TRACING_SERVICE_NAME", "event-tracking"),
			Endpoint:    common.GetEnv("TRACING_ENDPOINT", "localhost:4318"),
			UseGRPC:     common.GetEnvBool("TRACING_USE_GRPC", false),
			SampleRatio: common.GetEnvFloat64("TRACING_SAMPLE_RATIO", 1.0),
		},
		Sentry: SentryConfig{
			Enable:           common.GetEnvBool("SENTRY_ENABLE", false),
			DSN:              common.GetEnv("SENTRY_DSN", ""),
			SampleRate:       common.GetEnvFloat64("SENTRY_SAMPLE_RATE", 1.0),
			TracesSampleRate: common.GetEnvFloat64("SENTRY_TRACES_SAMPLE_RATE", 0.2),
			Debug:            common.GetEnvBool("SENTRY_DEBUG", false),
		},
	}
}
