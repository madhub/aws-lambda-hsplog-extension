package telemetryApi

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
	"github.com/philips-software/go-hsdp-api/logging"
)

// const maxLogBatchSize int = 25
const applicationName string = "DicomStore"
const resourceType string = "LogEvent"
const serverName string = "AWS Lambda"
const user string = "lambdauser"
const eventId string = "1"
const pipeSeparator string = "|"
const dotSeparator string = "."
const eventType string = "function"
const timeMapKey string = "time"
const typeMapKey string = "type"
const recordMapKey string = "record"
const customLogEventString string = "CustomLogEvent"

// if log record type is 'function' and message contain 'CustomLogEvent' string
// this function process the message & sends to HSP logging service
// if it is 'function
func StoreLogs(loggingClient logging.Client, lambdaFunctionName string, lambdaFunctionVersion string, logEntries []interface{}) (*logging.StoreResponse, error) {

	hsplogEvents := make([]logging.Resource, 0)

	for logEntry := range logEntries {
		aLogEnt := logEntries[logEntry]
		myMap := aLogEnt.(map[string]interface{})
		timeRecord := myMap[timeMapKey]       // get time from the event
		evtType := myMap[typeMapKey].(string) // get event type from the event
		// process function events , all other events will goto cloudwatch such as events genarated by platform & extension
		if strings.Contains(evtType, eventType) {
			logMessage := myMap[recordMapKey].(string) // get record from the event
			if strings.Contains(logMessage, customLogEventString) {
				var logResource, _ = createLoggingResource(logMessage, timeRecord.(string), lambdaFunctionName, lambdaFunctionVersion)
				hsplogEvents = append(hsplogEvents, logResource)
			}
		}

	}

	if len(hsplogEvents) > 0 {
		return loggingClient.StoreResources(hsplogEvents, len(hsplogEvents))
	}
	return nil, nil
}

// creates a HSP LogEvent record from the message
func createLoggingResource(hsdpLogFormattedMessage string, timeStamp string, lambdaFunctionName string, lambdaFunctionVersion string) (logging.Resource, error) {
	// hsdp logformat.
	//Extract information from HSPD console format
	// <<LogCategory>>.<<Severity>>|CustomLogEvent|{TransactionId}|{TraceId}|{spanId}|{ComponentName}|<message> .
	items := strings.Split(hsdpLogFormattedMessage, pipeSeparator)
	subItems := strings.Split(items[0], dotSeparator)

	var logResource = logging.Resource{
		ID:                  uuid.New().String(),
		ResourceType:        resourceType,
		ApplicationName:     applicationName,
		EventID:             eventId,
		Category:            subItems[0],
		Component:           items[5],
		TransactionID:       items[2],
		ServiceName:         lambdaFunctionName,
		ApplicationInstance: items[5],
		ApplicationVersion:  lambdaFunctionVersion,
		OriginatingUser:     user,
		ServerName:          serverName,
		LogTime:             timeStamp,
		Severity:            subItems[1],
		TraceID:             items[3],
		SpanID:              items[4],
		LogData: logging.LogData{
			Message: base64.StdEncoding.EncodeToString([]byte(items[6])),
		},
	}
	return logResource, nil
}

// split a slice into uniform chunks
func CreateLogBatchChunks(slice []interface{}, chunkSize int) [][]interface{} {
	var chunks [][]interface{}
	for {
		if len(slice) == 0 {
			break
		}

		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}
