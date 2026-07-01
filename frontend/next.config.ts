import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Pin the workspace root to this project so Next does not pick a parent
  // lockfile when inferring the root for output file tracing.
  turbopack: {
    root: __dirname,
  },
};

export default nextConfig;
