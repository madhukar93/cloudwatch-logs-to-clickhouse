const http = require("http");
const https = require("https");

// ... other parts of your code ...

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

exports.handler = async (event, context) => {
  const url = process.env.TARGET_HOSTNAME;

  if (!url) {
    throw new Error("TARGET_URL environment variable is not set.");
  }

  const startTime = event.startTime();
  const response = await httpRequest(url);
  const completionTime = Date.now() - startTime;

  // Parse the JSON response
  let responseData;
  try {
    responseData = JSON.parse(response);
  } catch (error) {
    console.error("Failed to parse JSON response:", error);
  }

  console.log({
    apiTimestamp: new Date().toISOString(),
    apiCallCompletionTime: completionTime,
    apiUrl: url,
    responseStatusCode: responseData?.statusCode,
    thirdPartyName: event.body.third_party_name,
    tenantId: event.tenantId,
    entityType: event.body.entityType,
    entityId: event.body.entityId,

    // caller contains information about the invokation context like what lambda function was invoked,
    // what was the event, etc as a nested json object
    caller: {
      // maybe add some other useful information here from the api gateway event
      // avoiding nesting beyond 1 level
      lambdaFunctionName: context.lambdaFunctionName,
      lambdaFunctionVersion: context.lambdaFunctionVersion,
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
