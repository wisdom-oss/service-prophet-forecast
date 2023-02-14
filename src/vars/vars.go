// Package vars contains all globally used variables and their default values. Furthermore,
// the package also contains all internal errors since they are just variables to golang
package vars

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/qustavo/dotsql"
	"microservice/structs"
)

// ServiceName is the name of the service which is used for identifying it in the gateway
// TODO: Change the service name and remove the TODO comment
const ServiceName = "template-service"

// ===== Required Setting Variables =====
var (
	// APIGatewayHost contains the IP address or hostname of the used Kong API Gateway
	APIGatewayHost string

	// ServiceRoutePath is the path under which the instance of the microservice shall be reachable via the Kong API
	// Gateway
	ServiceRoutePath string

	// DatabaseHost specifies the host on which the postgres database runs on
	DatabaseHost string

	// DatabaseUser is the username of the postgres user accessing the database
	DatabaseUser string

	// DatabaseUserPassword is the password of the user accessing the database
	DatabaseUserPassword string

	// RedisAddress specifies to which redis database the microservice shall connect to to check for the queries
	RedisAddress string
)

// ===== Optional Setting Variables =====
var (
	// ListenPort is the port this microservice will listen on. It defaults to 8000
	ListenPort int = 8000

	// DatabasePort specifies on which port the database used listens on
	DatabasePort int = 5432

	// ScopeConfigurationPath specifies from where the service should read the configuration of the needed access scope
	ScopeConfigurationPath string = "/res/scope.json"

	// APIGatewayPort contains the port on which the admin api of the Kong API Gateway listens
	APIGatewayPort int = 8001

	// QueryFilePath specifies from where the service shall load the sql queries
	QueryFilePath string = "/res/queries.sql"
)

// ===== Globally used variables =====

// PostgresConnection is a connection shared throughout the service
var PostgresConnection *sql.DB

// ScopeConfiguration containing the information about the scope needed to access this service
var ScopeConfiguration *structs.ScopeInformation

// SqlQueries contains all loaded SQL queries for the project
var SqlQueries *dotsql.DotSql

// HttpLogger is a logger for routes and similar functions
var HttpLogger zerolog.Logger

// TemporaryDataDirectory stores the path for the temporary file directory
var TemporaryDataDirectory string

// RedisClient holds the connection to the redis server which caches the forecasts
var RedisClient *redis.Client
