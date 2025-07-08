package configs

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	ES7 "github.com/elastic/go-elasticsearch/v7"
)

type ElasticSearchConfig struct {
	Client *ES7.Client
}

var elasticInstance *ElasticSearchConfig
var elasticOnce sync.Once

func GetElasticSearch() *ES7.Client {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic setup elastic search, ", r)
		}
	}()

	elasticOnce.Do(func() {
		log.Println("Initialize ElasticSearch connection...")

		esConfig, err := getElasticConfig()
		if err != nil {
			panic(err.Error())
		}

		es7, err := ES7.NewClient(esConfig)
		if err != nil {
			panic(err.Error())
		}

		elasticInstance = &ElasticSearchConfig{
			Client: es7,
		}

		log.Printf("Success Initialize ElasticSearch database connection")
	})

	return elasticInstance.Client
}

// RabbitHost
type elasticHost struct {
	Address  string
	Username string
	Password string
	Port     string
}

func getElasticConfig() (config ES7.Config, err error) {
	var idx int
	configsData := make(map[int]elasticHost)
	configsData[0] = elasticHost{
		Address:  os.Getenv("ELASTIC_HOST"),
		Username: os.Getenv("ELASTIC_USERNAME"),
		Password: os.Getenv("ELASTIC_PASSWORD"),
		Port:     os.Getenv("ELASTIC_PORT"),
	}

	if len(configsData[0].Address) == 0 {
		err = errors.New("environment variable for ELASTIC_HOST is not found")
	} else if len(configsData[0].Username) == 0 {
		err = errors.New("environment variable for ELASTIC_USERNAME is not found")
	}

	idx = 1
	for {

		value := os.Getenv("ELASTIC_HOST" + strconv.Itoa(idx))
		if len(value) == 0 {
			break
		}

		configsData[idx] = elasticHost{
			Address:  os.Getenv("ELASTIC_HOST" + strconv.Itoa(idx)),
			Username: os.Getenv("ELASTIC_USERNAME" + strconv.Itoa(idx)),
			Password: os.Getenv("ELASTIC_PASSWORD" + strconv.Itoa(idx)),
			Port:     os.Getenv("ELASTIC_PORT" + strconv.Itoa(idx)),
		}

		idx++
	}

	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(len(configsData))
	configData := configsData[random]

	esConfig := ES7.Config{
		Addresses: []string{
			configData.Address + ":" + configData.Port,
		},
		Username: configData.Username,
		Password: configData.Password,
		// CloudID:   "",
		// APIKey:    "",
		// Transport: nil,
		// Logger:    nil,
	}

	return esConfig, err
}
