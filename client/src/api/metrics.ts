import { Request, Response } from "express";
import { config } from "../config.js";

export async function handlerReadiness(
  _: Request,
  res: Response,
): Promise<void> {
  const time = `Request received at: ${new Date()}\n`;
  res.set("Content-Type", "text/plain; charset=utf-8");
  res.send(`${time}OK`);
}

export async function handlerMetrics(_: Request, res: Response): Promise<void> {
  res.set("Content-Type", "text/html; charset=utf-8");
  res.send(`<html>
  <body>
    <h1>denethor client admin</h1>
    <p>client has been visited ${config.client.serverHits} times</p>
  </body>
</html>
`);
}
