package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func awsSession(localstackURL string) *session.Session {
	c := &aws.Config{
		Endpoint:   aws.String(localstackURL),
		Region:     aws.String("us-east-1"),
		DisableSSL: aws.Bool(true),
	}
	return session.Must(session.NewSession(c))
}

type lambdaParams struct {
	functionName string
	zipFile      []byte
	environment  *lambda.Environment
}

func NewNodeJsLambda(params lambdaParams, session *session.Session) *lambda.FunctionConfiguration {
	client := lambda.New(session)
	lambda, err := client.CreateFunction(&lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			ZipFile: params.zipFile,
		},
		// Duummy value
		// https://docs.localstack.cloud/user-guide/aws/iam/#soft-mode
		Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
		FunctionName: aws.String(params.functionName),
		Handler:      aws.String("main"),
		MemorySize:   aws.Int64(128),
		PackageType:  aws.String("Zip"),
		Runtime:      aws.String("nodejs18.x"),
		Environment:  params.environment,
	})
	if err != nil {
		log.Fatalf("Failed to create lambda function: %s", err)
	}

	createCloudwatchLogGroupForLambda(lambda, session)
	return lambda
}

func logGroupName(lambdaFunctionName string) string {
	return fmt.Sprintf("/aws/lambda/%s", lambdaFunctionName)
}

func createCloudwatchLogGroupForLambda(lambda *lambda.FunctionConfiguration, session *session.Session) {
	logsvc := cloudwatchlogs.New(session)
	createLogGroupInput := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName(*lambda.FunctionName)),
	}

	if _, err := logsvc.CreateLogGroup(createLogGroupInput); err != nil {
		log.Fatalf("Failed to create log group: %s", err)
	}
	logGroupName := aws.String(logGroupName(*lambda.FunctionName))
	createLogStreamInput := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  logGroupName,
		LogStreamName: aws.String("clickhouse-log-stream"),
	}

	if _, err := logsvc.CreateLogStream(createLogStreamInput); err != nil {
		log.Fatalf("Failed to create log stream: %s", err)
	}

	// create subscription filter
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/SubscriptionFilters.html
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html
	pubsubSubscriptionInput := &cloudwatchlogs.PutSubscriptionFilterInput{
		DestinationArn: lambda.FunctionArn,
		FilterName:     aws.String("clickhouse-subscription-filter"),
		FilterPattern:  aws.String(""),
		LogGroupName:   logGroupName,
	}

	if _, err := logsvc.PutSubscriptionFilter(pubsubSubscriptionInput); err != nil {
		log.Fatalf("Failed to create subscription filter: %s", err)
	}
}

func printLambdaLogs(sess *session.Session, lambdaFunctionName string, nextToken *string) *string {
	svc := cloudwatchlogs.New(sess)
	// Get the latest log stream
	streams, err := svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName(lambdaFunctionName)),
		Descending:   aws.Bool(false),
		Limit:        aws.Int64(5),
		OrderBy:      aws.String("LastEventTime"),
	})
	if err != nil {
		log.Fatalf("Failed to describe log streams: %s", err)
	}
	if len(streams.LogStreams) == 0 {
		log.Fatalf("No log streams found for function: %s", lambdaFunctionName)
	}

	// if multiple streams are returned, use the first one because they should all have the same logs
	logStreamName := *streams.LogStreams[0].LogStreamName

	// Fetch the latest log events from the stream
	query := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName(lambdaFunctionName)),
		LogStreamName: aws.String(logStreamName),
		NextToken:     nextToken,
		StartFromHead: aws.Bool(true),
		StartTime:     aws.Int64(time.Now().UnixMilli()),
	}
	resp, err := svc.GetLogEvents(query)
	if err != nil {
		log.Fatalf("Failed to get log events: %s", err)
	}

	// Display the log events
	for _, event := range resp.Events {
		fmt.Printf("%s: %s\n", lambdaFunctionName, *event.Message)
	}

	// Set the next token for the subsequent request
	return nextToken
}

func pollLambdaLogs(session *session.Session, logsSinkFunction *lambda.FunctionConfiguration) {
	for {
		var nextToken *string
		nextToken = printLambdaLogs(session, *logsSinkFunction.FunctionName, nextToken)
		time.Sleep(5 * time.Second)
	}
}

func waitForLambdaToBeActive(lambdaSvc *lambda.Lambda, functionName string) {
	for {
		result, err := lambdaSvc.GetFunction(&lambda.GetFunctionInput{
			FunctionName: &functionName,
		})
		if err != nil {
			log.Fatalf("Failed to get function: %s", err)
		}

		log.Printf("Function %s state: %s", functionName, *result.Configuration.State)

		// Check state - for LocalStack, it might instantly be active.
		// But in a real-world AWS scenario, you'd check if the state is active.
		if *result.Configuration.State == lambda.StateActive {
			break
		}

		time.Sleep(5 * time.Second) // Adjust polling interval as needed
	}
}
