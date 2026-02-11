package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	CfgFile string
	Verbose int
)

const (
	ConsoleCmdPrefixKey = "consoleCmdPrefix"
	ConsoleCmdDirKey    = "consoleCmdDir"
	ConsoleCmdKey       = "consoleCmd"

	queueNamesKey       = "worker.queues"
	syncDataEntitiesKey = "sync-data.entities"

	resourceKey = "resource"

	rabbitmqConnectionStringKey = "rabbitmq.connection-string"
	rabbitmqServerKey           = "rabbitmq.server"
	rabbitmqPortKey             = "rabbitmq.port"
	rabbitmqUserKey             = "rabbitmq.user"
	rabbitmqPasswordKey         = "rabbitmq.password"

	postgresConnectionStringKey = "postgres.connection-string"
	postgresServerKey           = "postgres.server"
	postgresPortKey             = "postgres.port"
	postgresUserKey             = "postgres.user"
	postgresPasswordKey         = "postgres.password"
	postgresDatabaseKey         = "postgres.database"

	QueueDeclareRetryWait = time.Second * 30
)

func init() {
	// Bind environment variables to viper keys
	viper.BindEnv(rabbitmqServerKey, "RABBITMQ_SERVER")
	viper.BindEnv(rabbitmqPortKey, "RABBITMQ_CLIENT_PORT")
	viper.BindEnv(rabbitmqUserKey, "RABBITMQ_ADMIN_USER")
	viper.BindEnv(rabbitmqPasswordKey, "RABBITMQ_ADMIN_PW")

	viper.BindEnv(postgresServerKey, "POSTGRESQL_SERVER")
	viper.BindEnv(postgresPortKey, "POSTGRESQL_PORT")
	viper.BindEnv(postgresUserKey, "POSTGRESQL_USER")
	viper.BindEnv(postgresPasswordKey, "POSTGRESQL_PW")
	viper.BindEnv(postgresDatabaseKey, "POSTGRESQL_DATABASE")

	// count verbosity level from command line arguments
	Verbose = countVerbosityLevel(os.Args)
}

type Config struct {
	ConsoleCmdPrefix []string
	ConsoleCmdDir    string
	ConsoleCmd       []string

	QueueNames       []string
	SyncDataEntities []SyncEntity

	ResourceFilter string

	RabbitMq RabbitMQ
	Postgres Postgres
}

type RabbitMQ struct {
	ConnectionString string
	Server           string
	Port             int
	User             string
	Password         string
}

type Postgres struct {
	ConnectionString string
	Server           string
	Port             int
	User             string
	Password         string
	Database         string
}

func New() *Config {
	var syncConfigEntities []SyncEntity

	err := viper.UnmarshalKey(syncDataEntitiesKey, &syncConfigEntities)
	if err != nil {
		log.Panicf("Failed to parse sync-data.entities config: %s", err)
	}

	return &Config{
		ConsoleCmdPrefix: strings.Split(viper.GetString(ConsoleCmdPrefixKey), " "),
		ConsoleCmdDir:    viper.GetString(ConsoleCmdDirKey),
		ConsoleCmd:       strings.Split(viper.GetString(ConsoleCmdKey), " "),
		QueueNames:       viper.GetStringSlice(queueNamesKey),
		SyncDataEntities: syncConfigEntities,
		ResourceFilter:   viper.GetString(resourceKey),
		RabbitMq: RabbitMQ{
			ConnectionString: viper.GetString(rabbitmqConnectionStringKey),
			Server:           viper.GetString(rabbitmqServerKey),
			Port:             viper.GetInt(rabbitmqPortKey),
			User:             viper.GetString(rabbitmqUserKey),
			Password:         viper.GetString(rabbitmqPasswordKey),
		},
		Postgres: Postgres{
			ConnectionString: viper.GetString(postgresConnectionStringKey),
			Server:           viper.GetString(postgresServerKey),
			Port:             viper.GetInt(postgresPortKey),
			User:             viper.GetString(postgresUserKey),
			Password:         viper.GetString(postgresPasswordKey),
			Database:         viper.GetString(postgresDatabaseKey),
		},
	}
}

func (r *RabbitMQ) GetConnectionString() string {
	if r.ConnectionString != "" {
		return r.ConnectionString
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%d/", r.User, r.Password, r.Server, r.Port)
}

func (p *Postgres) GetConnectionString() string {
	if p.ConnectionString != "" {
		return p.ConnectionString
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", p.User, p.Password, p.Server, p.Port, p.Database)
}

func countVerbosityLevel(args []string) int {
	count := 0
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Handle -v, -vv, -vvv
			if strings.HasPrefix(arg, "-v") {
				for _, ch := range arg[1:] {
					if ch == 'v' {
						count++
					}
				}
			}
		}
	}
	return count
}
