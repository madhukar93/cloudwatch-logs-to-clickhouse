import { createClient } from "@clickhouse/client";
import { KinesisStreamHandler } from "aws-lambda";
import zlib from "zlib";

const clickhouse = createClient({
  host: process.env.CLICKHOUSE_HOST,
});

export const handler: KinesisStreamHandler = async (event, context) => {
  try {
    const records = event.Records.map((record) => {
      // Assume the Kinesis record data is base64 encoded data and then gzipped
      const payload = zlib
        .gunzipSync(Buffer.from(record.kinesis.data, "base64"))
        .toString("utf-8");
      return JSON.parse(payload);
    });

    /*
     console.log(JSON.stringify({
    apiTimestamp: new Date().toISOString(),
    apiCallCompletionTime: completionTime,
    apiUrl: url,
    // 0 when there's connection failure
    httpStatusCode: response.statusCode || 0,
    thirdPartyName: event.ThirdPartyName,
    tenantId: event.TenantId,
    operationCategory: event.OperationCategory,
    operationSubCategory: event.OperationCategory,
    entityType: event.EntityType,
    entityId: event.EntityId,

    // caller contains information about the invokation context like what lambda function was invoked,
    // what was the event, etc as a nested json object
    caller: {
      // maybe add some other useful information here from the api gateway event
      // avoiding nesting beyond 1 level
      lambdaFunctionName: context.functionName,
      lambdaFunctionVersion: context.functionVersion,
      lambdaFunctionArn: context.invokedFunctionArn,
      logGroupName: context.logGroupName,
      logStreamName: context.logStreamName,
      awsRequestId: context.awsRequestId,
    },
  }));

     */

    const values = records.map((record) => {
      api_timestamp: record.apiTimestamp;
      api_call_completion_time: record.apiCallCompletionTime;
      api_url: record.apiUrl;
      http_status_code: record.httpStatusCode;
      third_party_name: record.thirdPartyName;
      tenant_id: record.tenantId;
      operation_category: record.operationCategory;
      operation_sub_category: record.operationSubCategory;
      entity_type: record.entityType;
      entity_id: record.entityId;
      context: JSON.stringify(record.caller);
    });

    console.info({
      result: await clickhouse.insert({
        table: "api_logs",
        values: values,
        format: "JSONEachRow",
      }),
      context: {
        awsRequestId: context.awsRequestId,
        functionName: context.functionName,
        functionVersion: context.functionVersion,
        invokedFunctionArn: context.invokedFunctionArn,
      },
    });
  } catch (error) {
    // TODO: catch decode error, catch json parse error separately
    console.log(error);
  }
  // TODO: should a connection be created and destroyed every event, or is there a nicer way to tear it down?
  // await clickhouse.close();
};
