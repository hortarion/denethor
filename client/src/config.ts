type Config = {
  client: ClientConfig;
};

type ClientConfig = {
  port: number;
  serverHits: number;
};

process.loadEnvFile();

function envOrThrow(key: string) {
  const value = process.env[key];
  if (!value) {
    throw new Error(`Environment variable ${key} is not set`);
  }
  return value;
}

export const config: Config = {
  client: {
    port: Number(envOrThrow("PORT")),
    serverHits: 0,
  },
};
