import express, { NextFunction, Request, Response } from "express";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import { handlerInput } from "./api/input.js";
import { handlerReadiness } from "./api/metrics.js";
import { middlewareLogResponse, errorMiddleWare } from "./api/middleware.js";

async function main() {
  const __dirname = dirname(fileURLToPath(import.meta.url));

  const page = express();
  page.use(express.static(join(__dirname)));
  page.get("/", (req, res) => {
    res.sendFile(join(__dirname, "index.html"));
  });

  // Middleware
  page.use(middlewareLogResponse);
  page.use(express.json());

  // APIs
  page.get("/api/ping", (req: Request, res: Response, next: NextFunction) => {
    Promise.resolve(handlerReadiness(req, res)).catch(next);
  });
  page.post(
    "/api/handleInput",
    (req: Request, res: Response, next: NextFunction) => {
      Promise.resolve(handlerInput(req, res)).catch(next);
    },
  );

  // Error handling
  page.use(errorMiddleWare);

  // Listening to port
  page.listen(7070, () => {
    console.log(`Serving page at http://localhost:7070`);
  });
}

main().catch(console.error);
