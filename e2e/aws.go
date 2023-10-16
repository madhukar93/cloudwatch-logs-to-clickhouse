package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/tidwall/pretty"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/kinesis"
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
		Handler:      aws.String("index.handler"),
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
}

func createKinesisDataStream(session *session.Session) *kinesis.StreamDescription {
	kinesisSvc := kinesis.New(session)
	streamName := "api_logs_to_cloudwatch"
	shardCount := int64(1)

	_, err := kinesisSvc.CreateStream(&kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int64(shardCount),
	})

	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	log.Printf("Stream %s created successfully", streamName)

	output, err := kinesisSvc.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		log.Fatalf("Failed to describe Kinesis stream: %v", err)
	}

	return output.StreamDescription
}

func createCloudwatchSubscription(destinationArn string, lambdaFunctionName string, session *session.Session) {
	// create subscription filter
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/SubscriptionFilters.html
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html
	logSvc := cloudwatchlogs.New(session)
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/SubscriptionFilters.html
	pubsubSubscriptionInput := &cloudwatchlogs.PutSubscriptionFilterInput{
		DestinationArn: aws.String(destinationArn),
		FilterName:     aws.String("clickhouse-api-log"),
		// https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html
		FilterPattern: aws.String("{ ($.entityId != null) || ($.entityId != \"\") }"),
		// The name of the log group.
		LogGroupName: aws.String(logGroupName(lambdaFunctionName)),
		//   // The ARN of an IAM role that grants CloudWatch Logs permissions to deliver
		//   // ingested log events to the destination stream. You don't need to provide
		//   // the ARN when you are working with a logical destination for cross-account
		//   // delivery.
		RoleArn: aws.String("arn:aws:iam::123456789012:role/role-to-push-to-kinesis"),
	}

	if _, err := logSvc.PutSubscriptionFilter(pubsubSubscriptionInput); err != nil {
		log.Fatalf("Failed to create subscription filter: %s", err)
	} else {
		log.Printf("Subscription filter created successfully")
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

func viewKinesisRecords(streamName *string, session *session.Session) {
	// Create a new session and a Kinesis service client
	kinesisClient := kinesis.New(session)
	listShardsOutput, err := kinesisClient.ListShards(&kinesis.ListShardsInput{
		StreamName: streamName,
	})

	if err != nil {
		log.Fatalf("Failed to list shards: %v", err)
	}

	// Obtain a shard iterator
	for _, shard := range listShardsOutput.Shards {
		getShardIteratorOutput, err := kinesisClient.GetShardIterator(&kinesis.GetShardIteratorInput{
			StreamName:        streamName,
			ShardId:           shard.ShardId,
			ShardIteratorType: aws.String("TRIM_HORIZON"),
		})
		if err != nil {
			log.Fatalf("Failed to get shard iterator: %v", err)
		}
		shardIterator := getShardIteratorOutput.ShardIterator

		// Use the shard iterator to get records
		getRecordsOutput, err := kinesisClient.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: shardIterator,
		})
		if err != nil {
			log.Fatalf("Failed to get records: %v", err)
		}

		// Decode and print the event data
		for _, record := range getRecordsOutput.Records {
			// write record.Data to a file
			// os.WriteFile(fmt.Sprintf("kinesis-record-%d.txt", i), record.Data, 0644)
			// b64decoded, err := base64.StdEncoding.DecodeString(string(record.Data))
			// if err != nil {
			//   log.Printf("Failed to decode record data: %v", err)
			//   continue
			// }

			reader := bytes.NewReader(record.Data)
			gzipReader, err := gzip.NewReader(reader)
			if err != nil {
				log.Printf("Failed to create gzip reader: %v", err)
				continue
			}

			output, err := ioutil.ReadAll(gzipReader)
			if err != nil {
				log.Printf("Failed to read gzip data: %v", err)
				continue
			}

			fmt.Printf("Sequence Number: %s, Partition Key: %s, Data: %s\n",
				*record.SequenceNumber, *record.PartitionKey, pretty.Color(pretty.Pretty(output), nil))
		}
	}
}
