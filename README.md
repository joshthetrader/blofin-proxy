# Blofin CORS Proxy

A minimal Go server that proxies requests to Blofin API while adding CORS headers. This replaces expensive Netlify serverless functions with a cost-effective backend solution.

## Features

- ✅ **Pure CORS Proxy** - No credential storage or processing
- ✅ **Stateless** - No data persistence, maximum security
- ✅ **Fast** - No cold starts, persistent connections
- ✅ **Minimal** - Only Go standard library, ~6MB binary
- ✅ **Secure** - Forwards authentication headers without inspection

## Security Model

This proxy maintains your existing security architecture:

- **Credentials stay in browser** - localStorage only
- **Authentication happens client-side** - Web Crypto API signatures
- **Proxy is stateless** - No logging or storage of sensitive data
- **Direct WebSocket connections** - Real-time data bypasses proxy

## Local Development

```bash
# Run locally
go run main.go

# Test health check
curl http://localhost:8080/health

# Test proxy (with your actual headers)
curl -H "ACCESS-KEY: your-key" \
     -H "ACCESS-SIGN: your-signature" \
     http://localhost:8080/api/v1/market/tickers
```

## Deployment Options

### Option 1: Railway (Recommended - $5/month)

1. Push code to GitHub
2. Connect Railway to your repo
3. Deploy automatically
4. Get URL: `https://your-app.railway.app`

### Option 2: Render (Free tier available)

1. Push code to GitHub  
2. Connect Render to your repo
3. Uses `render.yaml` config
4. Get URL: `https://your-app.onrender.com`

### Option 3: DigitalOcean App Platform

1. Push code to GitHub
2. Create new App in DigitalOcean
3. Select your repo
4. Uses Dockerfile automatically

### Option 4: Docker (Any VPS)

```bash
# Build and run with Docker
docker build -t blofin-proxy .
docker run -p 8080:8080 blofin-proxy

# Or use docker-compose
docker-compose up -d
```

## Environment Variables

- `PORT` - Server port (default: 8080)
- `DEBUG` - Enable request logging (default: false)

## Frontend Integration

Update your frontend to use the deployed backend URL:

```javascript
// Replace this:
const restBase = '/blofin-api';

// With your deployed URL:
const restBase = 'https://your-backend-url.com/api';
```

## Cost Comparison

- **Before**: ~$36/month per active user (Netlify functions)
- **After**: $5/month total (handles unlimited users)
- **Savings**: 700x cost reduction for active users

## Monitoring

Health check endpoint: `GET /health`

Returns:
```json
{
  "status": "ok", 
  "timestamp": "2024-01-01T12:00:00Z"
}
```
