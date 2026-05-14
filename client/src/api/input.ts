import { Request, Response } from "express";

export async function handlerInput(req: Request, res: Response) {
  const body = req.body;
  var output = "";
  if (body["input"] === "hello") {
    output = "hello there";
  }
  res.header("Access-Control-Allow-Origin", "http://localhost:8080");
  res.send(output);
}
