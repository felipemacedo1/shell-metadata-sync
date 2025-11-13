import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export', // Static export for GitHub Pages
  basePath: process.env.NODE_ENV === 'production' ? '/dev-metadata-sync' : '',
  images: {
    unoptimized: true, // Required for static export
  },
  reactCompiler: true,
};

export default nextConfig;
