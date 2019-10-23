package utils

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
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
		Username:       os.Getenv("USERNAME_DB"),
		Password:       os.Getenv("PASSWORD_DB"),
		DatabaseType:   os.Getenv("DATABASE_TYPE"),
		DatabaseHost:   os.Getenv("DATABASE_HOST"),
		DatabasePort:   os.Getenv("DATABASE_PORT"),
		DatabaseName:   os.Getenv("DATABASE_NAME"),
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

		isDebug, err := strconv.ParseBool(os.Getenv("DATABASE_DEBUG"))
		if err != nil {
			panic(err)
		}

		/**
		 * NOTES: this will set connection lifetime in connection pool to 1 minute.
		 * 		  If the connection in the pool is idle > 1 min, Golang will close it
		 * 		  and will create new connection if #connections in the pool < pool max num
		 * 		  of connection. This to avoid invalid connection issue
		 */
		db.DB().SetConnMaxLifetime(time.Second * 60)
		db.SingularTable(true) // Set as singular table
		db.LogMode(isDebug)    // check database log mode

		dbInstance = &dbUtil{
			db: db,
		}
	})

	return dbInstance.db
}
