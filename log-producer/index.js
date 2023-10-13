const http = require("http");
const https = require("https");
const urllib = require("url");

function httpRequest(url) {
  return new Promise((resolve, reject) => {
    let client = url.startsWith("https://") ? https : http;

    client
      .get(url, (res) => {
        let data = "";

        // A chunk of data has been received.
        res.on("data", (chunk) => {
          data += chunk;
        });

        // The whole response has been received.
        res.on("end", () => {
          resolve({
            statusCode: res.statusCode,
            body: data,
          });
        });
      })
      .on("error", (err) => {
        reject(err);
      });
  });
}

// lambda function that makes an http request and logs
// API timestamp - allow range queries b/w such and such time
// API call completion time - for count, filtering
// URL - for count, filtering
// HTTP status code - count, filtering
// third party name - group by - count, filtering
// tenantId - customer
// Entity type
// Entity ID
// Operation category, operation subcategory
const mocks = {
  success: "success",
  internalServerError: "server-failure",
  delayedResponse: "latency",
  connectionReset: "connection-reset",
  notFound: "not-found",
  validationFailure: "validation-error",
};

const host = (() => {
  console.log("API_HOST: ", process.env.API_HOST);
  if (!process.env.API_HOST) {
    throw new Error("TARGET_URL environment variable is not set.");
  }
  return process.env.API_HOST;
})();

exports.handler = async (event, context) => {
  // base64 decode event.awslogs.data
  console.log("Received event:", JSON.stringify(event, null, 2));
  console.log("context:", JSON.stringify(context, null, 2));
  if (!event) {
    return;
  }

  const mock = mocks[event["MockScenario"]];

  console.log("Mock scenario: ", mock);

  if (!mock) {
    throw new Error("Mock scenario not found");
  }

  const url = urllib.format({
    protocol: "http",
    host,
    pathname: mock,
  });

  console.log("Making request to: ", url);

  const startTime = Date.now();
  let response;
  try {
    response = await httpRequest(url);
    console.log("Response: ", response);
  } catch (error) {
    console.log("Http request error: ", error);
  }
  const completionTime = Date.now() - startTime;

  console.log({
    apiTimestamp: new Date().toISOString(),
    apiCallCompletionTime: completionTime,
    apiUrl: url,
    // 0 when there's connection failure
    httpStatusCode: response.statusCode || 0,
    thirdPartyName: event.ThirdPartyName,
    tenantId: event.TenantId,
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
  });

  return {
    statusCode: 200,
    body: JSON.stringify(response),
  };
};
