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

type Dispatcher struct {
	//	httpClient       *http.Client
	hspLoggingClient *logging.Client
	postUri          string
	minBatchSize     int64
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
		hspLoggingClient: client,
		//		postUri:          dispatchPostUri,
		minBatchSize: dispatchMinBatchSize,
	}

}

func (d *Dispatcher) Dispatch(ctx context.Context, logEventsQueue *queue.Queue, force bool) {
	if !logEventsQueue.Empty() && (force || logEventsQueue.Len() >= d.minBatchSize) {
		l.Info("[dispatcher:Dispatch] Dispatching", logEventsQueue.Len(), "log events")
		logEntries, _ := logEventsQueue.Get(logEventsQueue.Len())
		// store log entries into HSP Logging service
		_, err := StoreLogs(*d.hspLoggingClient, logEntries)
		if err != nil {
			l.Error("[dispatcher:Dispatch] Failed to dispatch, returning to queue:", err)
			for logEntry := range logEntries {
				logEventsQueue.Put(logEntry)
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
