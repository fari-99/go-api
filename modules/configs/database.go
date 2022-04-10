package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type dbUtil struct {
	db *gorm.DB
}

const MySQLType = "MYSQL"
const PostgresType = "POSTGRES"

var dbInstance *dbUtil
var dbOnce sync.Once

type DatabaseConfig struct {
	Type string `json:"type"`

	Username       string `json:"username"`
	Password       string `json:"password"`
	DatabaseType   string `json:"database_type"`
	DatabaseHost   string `json:"database_host"`
	DatabasePort   string `json:"database_port"`
	DatabaseName   string `json:"database_name"`
	DatabaseConfig string `json:"database_config"`
}

func DatabaseBase(databaseType string) *DatabaseConfig {
	if !checkType(databaseType) {
		panic(fmt.Sprintf("database type [%s] not recognized, need configuration for that type", databaseType))
	}

	// default database configuration
	databaseConfig := DatabaseConfig{
		Type:           databaseType,
		Username:       os.Getenv(fmt.Sprintf("USERNAME_DB_%s", databaseType)),
		Password:       os.Getenv(fmt.Sprintf("PASSWORD_DB_%s", databaseType)),
		DatabaseType:   os.Getenv(fmt.Sprintf("DATABASE_TYPE_%s", databaseType)),
		DatabaseHost:   os.Getenv(fmt.Sprintf("DATABASE_HOST_%s", databaseType)),
		DatabasePort:   os.Getenv(fmt.Sprintf("DATABASE_PORT_%s", databaseType)),
		DatabaseName:   os.Getenv(fmt.Sprintf("DATABASE_NAME_%s", databaseType)),
		DatabaseConfig: os.Getenv(fmt.Sprintf("DATABASE_CONFIG_%s", databaseType)),
	}

	return &databaseConfig
}

func checkType(databaseType string) bool {
	switch databaseType {
	case MySQLType, PostgresType:
		return true
	default:
		return false
	}
}

func (base *DatabaseConfig) getConnection() string {
	var conn string
	if base.Type == MySQLType {
		conn = base.Username + ":" +
			base.Password + "@tcp(" +
			base.DatabaseHost + ":" +
			base.DatabasePort + ")/" +
			base.DatabaseName + "?" +
			base.DatabaseConfig
	} else if base.Type == PostgresType {
		conn = "host=" + base.DatabaseHost +
			" user=" + base.Username +
			" password=" + base.Password +
			" dbname=" + base.DatabaseName +
			" port=" + base.DatabasePort +
			" " + base.DatabaseConfig
	}

	return conn
}

func (base *DatabaseConfig) getLogger() logger.Interface {
	isDebug, _ := strconv.ParseBool(os.Getenv("DATABASE_DEBUG"))
	if !isDebug {
		return logger.Default
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			IgnoreRecordNotFoundError: false,       // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	return newLogger
}

func (base *DatabaseConfig) SetConnection() (*gorm.DB, error) {
	conn := base.getConnection()
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: base.getLogger(),
	}

	if base.Type == MySQLType {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN: conn,
		}), gormConfig)
		return db, err
	} else if base.Type == PostgresType {
		db, err := gorm.Open(postgres.New(postgres.Config{
			DSN:                  conn,
			PreferSimpleProtocol: true,
		}), gormConfig)
		return db, err
	}
	return nil, fmt.Errorf("database type [%s] not recognized, need configuration for that type", base.Type)
}

func (base *DatabaseConfig) GetMysqlConnection() *gorm.DB {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic setup database, ", r)
		}
	}()

	dbOnce.Do(func() {
		log.Println("Initialize database connection...")

		db, err := base.SetConnection()
		if err != nil {
			panic(err)
		}

		maxLifetime, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_CONNECTION_LIFETIME_MYSQL"), 10, 64)
		maxIdleConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_IDLE_CONNECTION_MYSQL"), 10, 64)
		maxOpenConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_OPEN_CONNECTION_MYSQL"), 10, 64)

		sqlDB, _ := db.DB()
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(maxLifetime)) // sets the maximum amount of time a connection may be reused.
		sqlDB.SetMaxIdleConns(int(maxIdleConn))                            // sets the maximum number of connections in the idle
		sqlDB.SetMaxOpenConns(int(maxOpenConn))                            // sets the maximum number of open connections to the database.

		dbInstance = &dbUtil{
			db: db,
		}
	})

	return dbInstance.db
}

func (base *DatabaseConfig) GetPostgresConnection() *gorm.DB {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic setup database, ", r)
		}
	}()

	dbOnce.Do(func() {
		log.Println("Initialize database connection...")

		db, err := base.SetConnection()
		if err != nil {
			panic(err)
		}

		maxLifetime, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_CONNECTION_LIFETIME_MYSQL"), 10, 64)
		maxIdleConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_IDLE_CONNECTION_MYSQL"), 10, 64)
		maxOpenConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_OPEN_CONNECTION_MYSQL"), 10, 64)

		sqlDB, _ := db.DB()
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(maxLifetime)) // sets the maximum amount of time a connection may be reused.
		sqlDB.SetMaxIdleConns(int(maxIdleConn))                            // sets the maximum number of connections in the idle
		sqlDB.SetMaxOpenConns(int(maxOpenConn))                            // sets the maximum number of open connections to the database.

		dbInstance = &dbUtil{
			db: db,
		}
	})

	return dbInstance.db
}
