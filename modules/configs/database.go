package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
)

type dbUtil struct {
	db *gorm.DB
}

var dbInstance *dbUtil
var dbOnce sync.Once

type DatabaseConfig struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	DatabaseType   string `json:"database_type"`
	DatabaseHost   string `json:"database_host"`
	DatabasePort   string `json:"database_port"`
	DatabaseName   string `json:"database_name"`
	DatabaseConfig string `json:"database_config"`
}

func DatabaseBase() *DatabaseConfig {
	// default database configuration
	databaseConfig := DatabaseConfig{
		Username:       os.Getenv("USERNAME_DB_MYSQL"),
		Password:       os.Getenv("PASSWORD_DB_MYSQL"),
		DatabaseType:   os.Getenv("DATABASE_TYPE_MYSQL"),
		DatabaseHost:   os.Getenv("DATABASE_HOST_MYSQL"),
		DatabasePort:   os.Getenv("DATABASE_PORT_MYSQL"),
		DatabaseName:   os.Getenv("DATABASE_NAME_MYSQL"),
		DatabaseConfig: "charset=utf8&parseTime=True&loc=Local",
	}

	return &databaseConfig
}

func (base *DatabaseConfig) GetConnection() string {
	conn := base.Username + ":" +
		base.Password + "@tcp(" +
		base.DatabaseHost + ":" +
		base.DatabasePort + ")/" +
		base.DatabaseName + "?" +
		base.DatabaseConfig

	return conn
}

func (base *DatabaseConfig) SetConnection() (*gorm.DB, error) {
	db, err := gorm.Open(base.DatabaseType, base.GetConnection())
	return db, err
}

func (base *DatabaseConfig) GetDBConnection() *gorm.DB {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic setup database, ", r)
		}
	}()

	dbOnce.Do(func() {
		log.Println("Initialize database connection...")

		databaseConfig := DatabaseBase()
		db, err := databaseConfig.SetConnection()

		if err != nil {
			panic(err)
		}

		isDebug, _ := strconv.ParseBool(os.Getenv("DATABASE_DEBUG"))
		maxLifetime, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_CONNECTION_LIFETIME_MYSQL"), 10, 64)
		maxIdleConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_IDLE_CONNECTION_MYSQL"), 10, 64)
		maxOpenConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_OPEN_CONNECTION_MYSQL"), 10, 64)

		db.DB().SetConnMaxLifetime(time.Second * time.Duration(maxLifetime)) // sets the maximum amount of time a connection may be reused.
		db.DB().SetMaxIdleConns(int(maxIdleConn))                            // sets the maximum number of connections in the idle
		db.DB().SetMaxOpenConns(int(maxOpenConn))                            // sets the maximum number of open connections to the database.
		db.SingularTable(true)                                               // Set as singular table
		db.LogMode(isDebug)                                                  // check database log mode

		dbInstance = &dbUtil{
			db: db,
		}
	})

	return dbInstance.db
}
