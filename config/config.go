package config

import (
	"fmt"
	"os"
	"strings"

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

func GetRabbitMQConnectionString() string {
	rabbitmqConnStr := viper.GetString(rabbitmqConnectionStringKey)
	if rabbitmqConnStr != "" {
		return rabbitmqConnStr
	}

	rabbitmqServer := viper.GetString(rabbitmqServerKey)
	rabbitmqPort := viper.GetInt(rabbitmqPortKey)
	rabbitmqUser := viper.GetString(rabbitmqUserKey)
	rabbitmqPassword := viper.GetString(rabbitmqPasswordKey)

	return fmt.Sprintf("amqp://%s:%s@%s:%d/", rabbitmqUser, rabbitmqPassword, rabbitmqServer, rabbitmqPort)
}

func GetPostgresConnectionString() string {
	postgresConnStr := viper.GetString(postgresConnectionStringKey)
	if postgresConnStr != "" {
		return postgresConnStr
	}

	postgresServer := viper.GetString(postgresServerKey)
	postgresPort := viper.GetInt(postgresPortKey)
	postgresUser := viper.GetString(postgresUserKey)
	postgresPassword := viper.GetString(postgresPasswordKey)
	postgresDatabase := viper.GetString(postgresDatabaseKey)

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", postgresUser, postgresPassword, postgresServer, postgresPort, postgresDatabase)
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
