import { createClient } from "@clickhouse/client";
import { KinesisStreamHandler } from "aws-lambda";

const clickhouse = createClient({
  host: process.env.CLICKHOUSE_HOST,
});

export const handler: KinesisStreamHandler = async (event, context) => {
  try {
    const records = event.Records.map((record) => {
      // Assume the Kinesis record data is base64 encoded
      const payload = Buffer.from(record.kinesis.data, "base64").toString(
        "utf-8"
      );
      return JSON.parse(payload);
    });

    // TODO: handle upsert and idempotence
    console.info({
      result: await clickhouse.insert({
        table: "api_logs",
        values: records,
        format: "JSONEachRow",
      }),
      context: {
        awsRequestId: context.awsRequestId,
        functionName: context.functionName,
        functionVersion: context.functionVersion,
        invokedFunctionArn: context.invokedFunctionArn,
      },
    });
    // TODO: should a connection be created and destroyed every event, or is there a nicer way to tear it down?
    // await clickhouse.close();
  } catch (error) {
    // TODO: catch decode error, catch json parse error separately
    console.log(error);
  }
};
