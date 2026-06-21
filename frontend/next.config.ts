import type { NextConfig } from "next";

const API_UPSTREAM = process.env.API_UPSTREAM_URL || "http://localhost:8080"

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/v1/:path*",
        destination: `${API_UPSTREAM}/v1/:path*`,
      },
    ]
  },
};

export default nextConfig;
