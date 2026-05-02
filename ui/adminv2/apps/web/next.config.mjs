/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    cpus: 4,
  },
  transpilePackages: ["@workspace/ui"],

  output: 'export',
  distDir: 'out',
  images: {
    unoptimized: true,
  },

  // When running 'npm run dev', the frontend runs on localhost:3000
  // and the backend runs on localhost:9090. This rule forwards /api requests to Go
  async rewrites() {
    // Note: Rewrites are ignored during 'npm run build' when output: 'export is set'.
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/api/:path*",
      },
    ];
  },
};

export default nextConfig
