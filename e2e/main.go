package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

/*
setup up lambda functions that ship logs to cloudwatch logs which are pushed to kinesis firehouse and read
on the other end by a lambda function that
*/

func handleKill() int {
	// This defer will be called when the function exits.
	defer fmt.Println("Deferred in run")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	sig := <-sigs
	fmt.Println()
	fmt.Println(sig)

	return 0
}

func main() {
	// clickhouseContainer := startClickHouseContainer()
	// defer clickhouseContainer.Terminate(context.Background())

	// var err error
	// clickhouseURL, err := clickhouseContainer.Endpoint(context.Background(), "8123/tcp")
	// if err != nil {
	//   log.Fatalf("Failed to get clickhouse endpoint: %s", err)
	// }

	localstackContainer := startlocalStackContainer()
	defer localstackContainer.Terminate(context.Background())

	localStackURL, err := localstackContainer.PortEndpoint(context.Background(), "4566", "")
	if err != nil {
		log.Fatalf("Failed to get localstack endpoint: %s", err)
	}

	session := awsSession("http://" + localStackURL)
	lambdaSvc := lambda.New(session)

	wiremock := startWiremockContainer()
	defer wiremock.Terminate(context.Background())

	wiremockURL, err := wiremock.PortEndpoint(context.Background(), "8080", "")

	if err != nil {
		log.Fatalf("Failed to get wiremock endpoint: %s", err)
	}

	setupStubs(wiremockURL)

	logProducerZipFile, err := os.ReadFile("../dist/log-producer.zip")

	if err != nil {
		log.Fatalf("Failed to read log-producer.zip: %s", err)
	}

	logProducerParams := lambdaParams{
		functionName: "log-producer",
		zipFile:      logProducerZipFile,
		environment: &lambda.Environment{
			Variables: map[string]*string{
				"WIREMOCK_URL": aws.String(wiremockURL),
			},
		},
	}
	logsProducerFunction := NewNodeJsLambda(logProducerParams, session)

	waitForLambdaToBeActive(lambdaSvc, *logsProducerFunction.FunctionName)

	log.Printf("Invoking %s", *logsProducerFunction.FunctionName)

	ps, err := payloads()
	if err != nil {
		log.Fatalf("Failed to get events: %s", err)
	}

	for _, p := range ps {
		lambdaSvc.Invoke(&lambda.InvokeInput{
			FunctionName: logsProducerFunction.FunctionName,
			Payload:      p,
		})
	}

	go pollLambdaLogs(session, logsProducerFunction)

	// logSinkZipFile, err := os.ReadFile("log-sink.zip")
	// logsSinkParams := lambdaParams{
	//   functionName: "log-sink",
	//   zipFile:      logSinkZipFile,
	//   environment: &lambda.Environment{
	//     Variables: map[string]*string{
	//       "CLICKHOUSE_URL": aws.String(clickhouseURL),
	//     },
	//   },
	// }
	// logsSinkFunction := NewNodeJsLambda(logsSinkParams, session)
	// lambdaSvc.Invoke(&lambda.InvokeInput{
	//   FunctionName: logsSinkFunction.FunctionName,
	//   Payload:      []byte(`{"key1":"value1", "key2":"value2", "key3":"value3"}`),
	// },
	// )
	// go pollLambdaLogs(session, logsSinkFunction)

	defer func() { os.Exit(handleKill()) }()
}
