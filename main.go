// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

/**

Notes:

- 	This is a simple example extension to help you start exploring the Lambda Telemetry API.
	This code is intended for eduational purposes only, it is not intended to run in production environments as-is.
	Use it as a reference only, at your own discretion, after you tested it thoroughly.

- 	Because of the asynchronous nature of the system, it is possible that telemetry for one invoke will be
	processed during the next invoke slice. Likewise, it is possible that telemetry for the last invoke will
	be processed during the SHUTDOWN event.

*/

package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/madhub/aws-lambda-hsplog-extension/extensionApi"
	"github.com/madhub/aws-lambda-hsplog-extension/telemetryApi"

	log "github.com/sirupsen/logrus"
)

var l = log.WithFields(log.Fields{"pkg": "main"})

func main() {
	// enable/disable verbose logging from environment variable
	logLevel := log.ErrorLevel
	verboseLogging := os.Getenv("ENABLE_VERBOSE_LOGGING")
	if len(verboseLogging) == 0 {
		logLevel = log.InfoLevel
	}

	l.Logger.SetLevel(logLevel)

	l.Info("[main] Starting the Telemetry API extension")
	extensionName := path.Base(os.Args[0])

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		l.Info("[main] Received", s)
		l.Info("[main] Exiting")
	}()

	// Step 1 - Register the extension with Extensions API
	l.Info("[main] Registering extension")
	extensionApiClient := extensionApi.NewClient()
	extensionId, err := extensionApiClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	l.Info("[main] Registation success with extensionId", extensionId)

	// Step 2 - Start the local http listener which will receive data from Telemetry API
	l.Info("[main] Starting the Telemetry listener")
	telemetryListener := telemetryApi.NewTelemetryApiListener()
	telemetryListenerUri, err := telemetryListener.Start()
	if err != nil {
		panic(err)
	}

	// Step 3 - Subscribe the listener to Telemetry API
	l.Info("[main] Subscribing to the Telemetry API")
	telemetryApiClient := telemetryApi.NewClient()
	_, err = telemetryApiClient.Subscribe(ctx, extensionId, telemetryListenerUri)
	if err != nil {
		panic(err)
	}
	l.Info("[main] Subscription success")

	dispatcher := telemetryApi.NewDispatcher()

	// Will block until invoke or shutdown event is received or cancelled via the context.
	for {
		select {
		case <-ctx.Done():
			return
		default:
			l.Info("[main] Waiting for next event...")

			// This is a blocking action
			res, err := extensionApiClient.NextEvent(ctx)
			if err != nil {
				l.Error("[main] Exiting. Error:", err)
				return
			}

			// Dispatching log events from previous invocations
			dispatcher.Dispatch(ctx, telemetryListener.LogEventsQueue, false)

			l.Info("[main] Received event")

			if res.EventType == extensionApi.Invoke {
				handleInvoke(res)
			} else if res.EventType == extensionApi.Shutdown {
				// Dispatch all remaining telemetry, handle shutdown
				dispatcher.Dispatch(ctx, telemetryListener.LogEventsQueue, true)
				handleShutdown(res)
				return
			}
		}
	}
}

func handleInvoke(r *extensionApi.NextEventResponse) {
	l.Info("[handleInvoke]")
}

func handleShutdown(r *extensionApi.NextEventResponse) {
	l.Info("[handleShutdown]")
}
