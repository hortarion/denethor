import express, { NextFunction, Request, Response } from "express";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import { config } from "./config.js";
import { handlerInput } from "./api/input.js";
import { handlerMetrics, handlerReadiness } from "./api/metrics.js";
import {
  errorMiddleWare,
  middlewareLogResponse,
  middlewareMetricsInc,
} from "./api/middleware.js";

async function main() {
  const __dirname = dirname(fileURLToPath(import.meta.url));
  // starting client
  console.log("starting client...");
  const client = express();
  client.use(express.static(join(__dirname)));

  // Middleware
  client.use(middlewareLogResponse);
  client.use(express.json());
  client.use("/", middlewareMetricsInc, express.static("./public/app"));

  // APIs
  client.get("/", (req: Request, res: Response) => {
    res.sendFile(join(__dirname, "index.html"));
  });
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
  client.post(
    "/api/handleInput",
    (req: Request, res: Response, next: NextFunction) => {
      Promise.resolve(handlerInput(req, res)).catch(next);
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
