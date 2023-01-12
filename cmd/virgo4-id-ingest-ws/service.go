package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// number of times to retry a message put before giving up and terminating
var sendRetries = uint(3)

// ServiceContext contains common data used by all handlers
type ServiceContext struct {
	config *ServiceConfig
	aws    awssqs.AWS_SQS
	queue  awssqs.QueueHandle
}

type ServiceResponse struct {
	Message string `json:"msg"`
}

// InitializeService will initialize the service context based on the config parameters.
func InitializeService(cfg *ServiceConfig) *ServiceContext {
	log.Printf("initializing service")

	// load our AWS_SQS helper object
	aws, err := awssqs.NewAwsSqs(awssqs.AwsSqsConfig{MessageBucketName: cfg.MessageBucketName})
	fatalIfError(err)

	// our SQS output queue
	outQueueHandle, err := aws.QueueHandle(cfg.OutQueueName)
	fatalIfError(err)

	svc := ServiceContext{
		config: cfg,
		aws:    aws,
		queue:  outQueueHandle,
	}

	return &svc
}

// IgnoreHandler is a dummy to handle certain browser requests without warnings (e.g. favicons)
func (svc *ServiceContext) IgnoreHandler(c *gin.Context) {
}

// VersionHandler reports the version of the service
func (svc *ServiceContext) VersionHandler(c *gin.Context) {
	vMap := make(map[string]string)
	vMap["build"] = Version()
	c.JSON(http.StatusOK, vMap)
}

// HealthCheckHandler reports the health of the serivce
func (svc *ServiceContext) HealthCheckHandler(c *gin.Context) {

	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}

	healthy := true
	hcDB := hcResp{Healthy: healthy}
	hcMap := make(map[string]hcResp)
	hcMap["service"] = hcDB

	hcStatus := http.StatusOK
	if healthy == false {
		hcStatus = http.StatusInternalServerError
	}

	c.JSON(hcStatus, hcMap)
}

func (svc *ServiceContext) IdIngestHandler(c *gin.Context) {

	id := c.Param("id")

	// send to outbound queue
	err := svc.queueOutbound(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ServiceResponse{err.Error()})
		return
	}

	// all good
	log.Printf("INFO: item id: %s queued", id)
	c.JSON(http.StatusOK, ServiceResponse{"OK"})
}

func (svc *ServiceContext) queueOutbound(id string) error {
	outbound := svc.constructMessage(id)
	messages := make([]awssqs.Message, 0, 1)
	messages = append(messages, *outbound)
	opStatus, err := svc.aws.BatchMessagePut(svc.queue, messages)
	if err != nil {
		// if an error we can handle, retry
		if err == awssqs.ErrOneOrMoreOperationsUnsuccessful {
			log.Printf("WARNING: item failed to send to the work queue, retrying...")

			// retry the failed item and bail out if we cannot retry
			err = svc.aws.MessagePutRetry(svc.queue, messages, opStatus, sendRetries)
		}
	}

	return err
}

// construct the outbound SQS message
func (svc *ServiceContext) constructMessage(id string) *awssqs.Message {

	attributes := make([]awssqs.Attribute, 0, 5)
	attributes = append(attributes, awssqs.Attribute{Name: awssqs.AttributeKeyRecordId, Value: id})
	attributes = append(attributes, awssqs.Attribute{Name: awssqs.AttributeKeyRecordSource, Value: svc.config.DataSourceName})
	attributes = append(attributes, awssqs.Attribute{Name: awssqs.AttributeKeyRecordOperation, Value: awssqs.AttributeValueRecordOperationUpdate})
	return &awssqs.Message{Attribs: attributes, Payload: []byte(id)}
}

//
// end of file
//
