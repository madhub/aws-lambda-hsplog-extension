package telemetryApi

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/philips-software/go-hsdp-api/logging"
)

// if log record type is 'function' and message contain 'CustomLogEvent' string
// this function process the message & sends to HSP logging service
// if it is 'function
func StoreLogs(loggingClient logging.Client, logEntries []interface{}) (*logging.StoreResponse, error) {

	hsplogEvents := make([]logging.Resource, 0)

	for logEntry := range logEntries {
		aLogEnt := logEntries[logEntry]
		myMap := aLogEnt.(map[string]interface{})
		timeRecord := myMap["time"]
		evtType := myMap["type"].(string)
		if strings.Contains(evtType, "function") {
			logMessage := myMap["record"].(string)
			if strings.Contains(logMessage, "CustomLogEvent") {
				var logResource, _ = CreateLoggingResource(logMessage, timeRecord.(string))
				hsplogEvents = append(hsplogEvents, logResource)
			}
		}

	}
	return loggingClient.StoreResources(hsplogEvents, len(hsplogEvents))
}

// creates a HSP LogEvent record from the message
func CreateLoggingResource(hsdpLogFormattedMessage string, timeStamp string) (logging.Resource, error) {
	// hsdp logformat.
	//Extract information from HSPD console format
	// <<LogCategory>>.<<Severity>>|CustomLogEvent|{TransactionId}|{TraceId}|{spanId}|{ComponentName}|<message> .
	items := strings.Split(hsdpLogFormattedMessage, "|")
	subItems := strings.Split(items[0], ".")

	lambdaFunctionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	lambdaFunctionVersion := os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")

	var logResource = logging.Resource{
		ID:                  uuid.New().String(),
		ResourceType:        "LogEvent",
		ApplicationName:     "DicomStore",
		EventID:             "1",
		Category:            subItems[0],
		Component:           items[5],
		TransactionID:       items[2],
		ServiceName:         lambdaFunctionName,
		ApplicationInstance: items[5],
		ApplicationVersion:  lambdaFunctionVersion,
		OriginatingUser:     "Ross",
		ServerName:          "AWS Lambda",
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
