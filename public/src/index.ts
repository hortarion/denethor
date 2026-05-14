import express from "express";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

async function main() {
  const __dirname = dirname(fileURLToPath(import.meta.url));
  const frontEndPort = 7070;

  const page = express();
  page.use(express.static(join(__dirname)));
  page.get("/", (req, res) => {
    res.sendFile(join(__dirname, "index.html"));
  });

  page.listen(frontEndPort, () => {
    console.log(`Serving page at http://localhost:${frontEndPort}`);
  });
}

main().catch(console.error);
