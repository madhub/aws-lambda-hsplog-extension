// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package telemetryApi

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/philips-software/go-hsdp-api/logging"
)

const maxLogBatchSize int = 25

type Dispatcher struct {
	//	httpClient       *http.Client
	hspLoggingClient      *logging.Client
	lambdaFunctionName    string
	lambdaFunctionVersion string
	postUri               string
	minBatchSize          int64
}

func NewDispatcher() *Dispatcher {
	// dispatchPostUri := os.Getenv("DISPATCH_POST_URI")
	// if len(dispatchPostUri) == 0 {
	// 	panic("dispatchPostUri undefined")
	// }
	hspLoggingBaseURL := os.Getenv("HSDP_LOGGING_BASE_URI")
	if len(hspLoggingBaseURL) == 0 {
		panic("hspLoggingBaseURL undefined")
	}
	productKey := os.Getenv("PRODUCT_KEY")
	if len(productKey) == 0 {
		panic("productKey undefined")
	}
	sharedKey := os.Getenv("SHARED_KEY")
	if len(sharedKey) == 0 {
		panic("sharedKey undefined")
	}
	sharedSecret := os.Getenv("SECRET_KEY")
	if len(sharedSecret) == 0 {
		panic("sharedSecret undefined")
	}

	lambdaFunctionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if len(lambdaFunctionName) == 0 {
		panic("lambdaFunctionName undefined")
	}

	dispatchMinBatchSize, err := strconv.ParseInt(os.Getenv("DISPATCH_MIN_BATCH_SIZE"), 0, 16)
	if err != nil {
		dispatchMinBatchSize = 1
	}
	client, err := logging.NewClient(http.DefaultClient, &logging.Config{
		SharedKey:    sharedKey,
		SharedSecret: sharedSecret,
		BaseURL:      hspLoggingBaseURL,
		ProductKey:   productKey})

	return &Dispatcher{
		//		httpClient:       &http.Client{},
		hspLoggingClient:      client,
		lambdaFunctionName:    lambdaFunctionName,
		lambdaFunctionVersion: "1.0",
		//		postUri:          dispatchPostUri,
		minBatchSize: dispatchMinBatchSize,
	}

}

func (d *Dispatcher) Dispatch(ctx context.Context, logEventsQueue *queue.Queue, force bool) {
	if !logEventsQueue.Empty() && (force || logEventsQueue.Len() >= d.minBatchSize) {

		l.Info("[dispatcher:Dispatch] Dispatching", logEventsQueue.Len(), "log events")
		logEntries, _ := logEventsQueue.Get(logEventsQueue.Len())
		// store logs in batches of 25 ( HSDP API log restriction )
		logsBatch := CreateLogBatchChunks(logEntries, maxLogBatchSize)
		for _, logEntriesBatch := range logsBatch {
			_, err := StoreLogs(*d.hspLoggingClient, d.lambdaFunctionName, d.lambdaFunctionVersion, logEntriesBatch)
			if err != nil {
				l.Error("[dispatcher:Dispatch] Failed to dispatch, returning to queue:", err)
				for logEntry := range logEntriesBatch {
					logEventsQueue.Put(logEntry)
				}
			}
		}

		// bodyBytes, _ := json.Marshal(logEntries)
		// req, err := http.NewRequestWithContext(ctx, "POST", d.postUri, bytes.NewBuffer(bodyBytes))
		// if err != nil {
		// 	panic(err)
		// }
		// _, err = d.httpClient.Do(req)
		// if err != nil {
		// 	l.Error("[dispatcher:Dispatch] Failed to dispatch, returning to queue:", err)
		// 	for logEntry := range logEntries {
		// 		logEventsQueue.Put(logEntry)
		// 	}
		// }
	}
}
