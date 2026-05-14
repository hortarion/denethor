import express, { NextFunction, Request, Response } from "express";
import { config } from "./config.js";
import { handlerMetrics, handlerReadiness } from "./api/metrics.js";
import {
  errorMiddleWare,
  middlewareLogResponse,
  middlewareMetricsInc,
} from "./api/middleware.js";

async function main() {
  // starting client
  console.log("starting client...");
  const client = express();

  // Middleware
  client.use(middlewareLogResponse);
  client.use(express.json());
  client.use("/", middlewareMetricsInc, express.static("./src/app"));

  // APIs
  client.get(
    "/api/health",
    (req: Request, res: Response, next: NextFunction) => {
      Promise.resolve(handlerReadiness(req, res)).catch(next);
    },
  );
  client.get(
    "/api/metrics",
    (req: Request, res: Response, next: NextFunction) => {
      Promise.resolve(handlerMetrics(req, res)).catch(next);
    },
  );

  // Error handling
  client.use(errorMiddleWare);

  // Listening to port
  client.listen(config.client.port, () => {
    console.log(`Server is running at http://localhost:${config.client.port}`);
  });
}

main().catch(console.error);
