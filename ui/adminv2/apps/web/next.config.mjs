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
};

export default nextConfig
