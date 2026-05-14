import { Request, Response } from "express";

export async function handlerReadiness(
  _: Request,
  res: Response,
): Promise<void> {
  const time = `Request received at: ${new Date()}\n`;
  res.set("Content-Type", "text/plain; charset=utf-8");
  res.send(`${time}OK`);
}
