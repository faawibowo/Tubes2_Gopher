import { createEnv } from "@t3-oss/env-nextjs";
import { z } from "zod";

export const env = createEnv({
  /*
   * Serverside Environment variables, not available on the client.
   * Will throw if you access these variables on the client.
   */
  server: {
    NODE_ENV: z
      .enum(["development", "test", "production"])
      .default("production"),
    SITE_URL: z.string().url(),
    BACKEND_URL: z.string().url(),
    NEXT_ANALYTICS_UMAMI_ID: z.string().default(""),
  },
  /*
   * Environment variables available on the client (and server).
   *
   * 💡 You'll get type errors if these are not prefixed with NEXT_PUBLIC_.
   */
  client: {
    NEXT_PUBLIC_BACKEND_HTTP: z.string().url(),
    NEXT_PUBLIC_BACKEND_WS: z.string().url(),
  },
  /*
   * Due to how Next.js bundles environment variables on Edge and Client,
   * we need to manually destructure them to make sure all are included in bundle.
   *
   * 💡 You'll get type errors if not all variables from `server` & `client` are included here.
   */
  experimental__runtimeEnv: {
    NEXT_PUBLIC_BACKEND_HTTP: process.env.NEXT_PUBLIC_BACKEND_HTTP,
    NEXT_PUBLIC_BACKEND_WS: process.env.NEXT_PUBLIC_BACKEND_WS,
  },
});
