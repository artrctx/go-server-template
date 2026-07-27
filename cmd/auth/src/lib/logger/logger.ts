import { NodeSDK } from "@opentelemetry/sdk-node";
import { OTLPLogExporter } from "@opentelemetry/exporter-logs-otlp-http";
import {
  BatchLogRecordProcessor,
  ConsoleLogRecordExporter,
  SimpleLogRecordProcessor,
} from "@opentelemetry/sdk-logs";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { env } from "../../env.js";
import { logs } from "@opentelemetry/api-logs";

const ServiceName = "shuffle-auth";
export const loggerSdk = new NodeSDK({
  resource: resourceFromAttributes({
    "service.name": ServiceName,
  }),

  logRecordProcessor:
    env.APP_ENV !== "development"
      ? new BatchLogRecordProcessor(
          new OTLPLogExporter({
            url: "https://us.i.posthog.com/i/v1/logs",
            headers: {
              Authorization: `Bearer ${env.POSTHOG_TOKEN}`,
            },
          }),
        )
      : new SimpleLogRecordProcessor(new ConsoleLogRecordExporter()),
});

export const logger = logs.getLogger(`${ServiceName}-logger`);
