package main

import (
	"log"
	"os"
	"strconv"
)

// ServiceConfig defines all the service configuration parameters
type ServiceConfig struct {
	ServicePort       int    // the service listen port
	OutQueueName      string // SQS queue name for outbound documents
	DataSourceName    string // the name to associate the data with. Each record has metadata showing this value
	MessageBucketName string // the bucket to use for large messages

	PayloadFormat string // the format of the generated payload, e.g "<id>%s</id>"
}

func ensureSet(env string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Printf("environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func ensureSetAndNonEmpty(env string) string {
	val := ensureSet(env)

	if val == "" {
		log.Printf("environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func envToInt(env string) int {

	number := ensureSetAndNonEmpty(env)
	n, err := strconv.Atoi(number)
	if err != nil {

		os.Exit(1)
	}
	return n
}

// LoadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func LoadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	cfg.ServicePort = envToInt("VIRGO4_ID_INGEST_WS_PORT")
	cfg.OutQueueName = ensureSetAndNonEmpty("VIRGO4_ID_INGEST_WS_OUT_QUEUE")
	cfg.DataSourceName = ensureSetAndNonEmpty("VIRGO4_ID_INGEST_WS_DATA_SOURCE")
	cfg.MessageBucketName = ensureSetAndNonEmpty("VIRGO4_SQS_MESSAGE_BUCKET")
	cfg.PayloadFormat = ensureSetAndNonEmpty("VIRGO4_ID_INGEST_WS_PAYLOAD_FORMAT")

	log.Printf("[CONFIG] ServicePort         = [%d]", cfg.ServicePort)
	log.Printf("[CONFIG] OutQueueName        = [%s]", cfg.OutQueueName)
	log.Printf("[CONFIG] DataSourceName      = [%s]", cfg.DataSourceName)
	log.Printf("[CONFIG] MessageBucketName   = [%s]", cfg.MessageBucketName)
	log.Printf("[CONFIG] PayloadFormat       = [%s]", cfg.PayloadFormat)

	return &cfg
}

//
// end of file
//
