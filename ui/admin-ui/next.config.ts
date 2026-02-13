import type { NextConfig } from "next";

const nextConfig: NextConfig = {
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
                destination: "http://localhost:9090/api/:path*",
            },
        ];
    },
};

export default nextConfig;
