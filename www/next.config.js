/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  distDir: process.env.NODE_ENV === 'development' ? 'out' : 'dist',
  images: { unoptimized: true },
  // for local development - (\\d{1,}) is for port number
  // async rewrites() {
  //   return [{ source: '/api/parse', destination: 'http://localhost:8080/api/parse' }]
  // }
}

module.exports = nextConfig
