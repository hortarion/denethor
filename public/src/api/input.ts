import { Request, Response } from "express";

export async function handlerInput(req: Request, res: Response) {
  const body = req.body;
  var output = "";
  if (body["input"] === "hello") {
    output = "hello there";
  }
  if (body["input"] === "ping") {
    output = "pong";
  }
  res.send(output);
}
