#!/bin/bash

# Blofin Proxy VPS Setup Script
# Run this on a fresh Ubuntu 22.04 VPS

set -e

echo "🚀 Setting up Blofin CORS Proxy on VPS..."

# Update system
echo "📦 Updating system packages..."
apt update && apt upgrade -y

# Install Docker
echo "🐳 Installing Docker..."
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
systemctl enable docker
systemctl start docker

# Install Docker Compose
echo "🔧 Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Install Nginx (for reverse proxy and SSL)
echo "🌐 Installing Nginx..."
apt install -y nginx

# Install Certbot (for SSL certificates)
echo "🔒 Installing Certbot..."
apt install -y certbot python3-certbot-nginx

# Create app directory
echo "📁 Setting up application directory..."
mkdir -p /opt/blofin-proxy
cd /opt/blofin-proxy

# Clone repository (you'll need to replace this URL)
echo "📥 Cloning repository..."
# git clone https://github.com/yourusername/blofin-proxy.git .
# For now, we'll create the files directly

# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  blofin-proxy:
    build: .
    ports:
      - "127.0.0.1:8080:8080"  # Only bind to localhost
    environment:
      - PORT=8080
      - DEBUG=false
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
EOF

# Create systemd service for auto-start
echo "⚙️ Creating systemd service..."
cat > /etc/systemd/system/blofin-proxy.service << 'EOF'
[Unit]
Description=Blofin CORS Proxy
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/blofin-proxy
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Enable the service
systemctl enable blofin-proxy.service

# Create Nginx configuration template
echo "🌐 Creating Nginx configuration template..."
cat > /etc/nginx/sites-available/blofin-proxy << 'EOF'
server {
    listen 80;
    server_name YOUR_DOMAIN_HERE;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # Proxy to Go backend
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers (backup - Go backend also sets these)
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, PATCH, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization, ACCESS-KEY, ACCESS-SIGN, ACCESS-TIMESTAMP, ACCESS-NONCE, ACCESS-PASSPHRASE, BROKER-ID";
    }

    # Health check endpoint
    location /health {
        proxy_pass http://127.0.0.1:8080/health;
        access_log off;
    }
}
EOF

# Create firewall rules
echo "🔥 Setting up firewall..."
ufw allow ssh
ufw allow 'Nginx Full'
ufw --force enable

echo "✅ VPS setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Copy your Go backend files to /opt/blofin-proxy/"
echo "2. Edit /etc/nginx/sites-available/blofin-proxy and replace YOUR_DOMAIN_HERE"
echo "3. Enable the Nginx site: ln -s /etc/nginx/sites-available/blofin-proxy /etc/nginx/sites-enabled/"
echo "4. Test Nginx config: nginx -t"
echo "5. Reload Nginx: systemctl reload nginx"
echo "6. Start the service: systemctl start blofin-proxy"
echo "7. Get SSL certificate: certbot --nginx -d yourdomain.com"
echo ""
echo "🌐 Your proxy will be available at: http://your-server-ip"
echo "📊 Monitor with: docker-compose logs -f"
echo "🔍 Health check: curl http://your-server-ip/health"
