# VPS Deployment Guide

Deploy your Blofin CORS proxy on a VPS for maximum cost efficiency and performance.

## 🏆 Why VPS?

- **Cheapest**: $3.50-6/month vs $36+/month per user
- **Fastest**: Always warm, no cold starts
- **Full Control**: Customize everything
- **Scalable**: Handle thousands of users on one server

## 🚀 Quick VPS Setup

### Step 1: Choose a VPS Provider

| Provider | Cost/Month | RAM | Storage | Bandwidth |
|----------|------------|-----|---------|-----------|
| **Vultr** | $3.50 | 512MB | 10GB | 0.5TB |
| **DigitalOcean** | $4.00 | 512MB | 10GB | 0.5TB |
| **Linode** | $5.00 | 1GB | 25GB | 1TB |
| **Hetzner** | €3.29 | 1GB | 20GB | 20TB |

**Recommendation**: Vultr $3.50/month is perfect for this proxy.

### Step 2: Create VPS

1. **Sign up** at your chosen provider
2. **Create server**:
   - OS: Ubuntu 22.04 LTS
   - Location: Closest to your users
   - Size: Smallest option (512MB RAM is enough)
3. **Note the IP address**

### Step 3: Automated Setup

```bash
# SSH into your VPS
ssh root@YOUR_SERVER_IP

# Download and run setup script
curl -fsSL https://raw.githubusercontent.com/yourusername/blofin-proxy/main/go-backend/vps-setup.sh -o setup.sh
chmod +x setup.sh
./setup.sh
```

### Step 4: Deploy Your Code

```bash
# On your VPS, copy the Go backend files
cd /opt/blofin-proxy

# Create the Go files (copy from your local go-backend folder)
# You can use scp, git clone, or copy-paste

# Create main.go
cat > main.go << 'EOF'
[Copy the contents of your go-backend/main.go here]
EOF

# Create go.mod  
cat > go.mod << 'EOF'
[Copy the contents of your go-backend/go.mod here]
EOF

# Create Dockerfile
cat > Dockerfile << 'EOF'
[Copy the contents of your go-backend/Dockerfile here]
EOF

# Start the service
systemctl start blofin-proxy
systemctl status blofin-proxy
```

### Step 5: Test It Works

```bash
# Test health endpoint
curl http://YOUR_SERVER_IP/health

# Should return: {"status":"ok","timestamp":"..."}
```

## 🌐 Optional: Add Custom Domain

### Step 1: Point Domain to VPS

In your domain registrar (Cloudflare, Namecheap, etc.):
```
A record: api.yourdomain.com → YOUR_SERVER_IP
```

### Step 2: Configure Nginx

```bash
# Edit Nginx config
nano /etc/nginx/sites-available/blofin-proxy

# Replace YOUR_DOMAIN_HERE with your actual domain
server_name api.yourdomain.com;

# Enable the site
ln -s /etc/nginx/sites-available/blofin-proxy /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

### Step 3: Add SSL Certificate

```bash
# Get free SSL certificate
certbot --nginx -d api.yourdomain.com

# Auto-renewal is set up automatically
```

## 🔧 Manual Setup (Alternative)

If you prefer manual setup:

### Step 1: Install Dependencies

```bash
# Update system
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

### Step 2: Deploy Application

```bash
# Create app directory
mkdir -p /opt/blofin-proxy
cd /opt/blofin-proxy

# Copy your files (main.go, go.mod, Dockerfile, docker-compose.yml)
# Then run:
docker-compose up -d
```

### Step 3: Set Up Reverse Proxy (Optional)

```bash
# Install Nginx
apt install -y nginx

# Create config
cat > /etc/nginx/sites-available/blofin-proxy << 'EOF'
server {
    listen 80;
    server_name YOUR_DOMAIN_OR_IP;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
EOF

# Enable site
ln -s /etc/nginx/sites-available/blofin-proxy /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

## 📊 Update Frontend Configuration

### Option 1: Use IP Address

```javascript
// src/config.js
const BACKEND_CONFIG = {
  PRODUCTION_URL: 'http://YOUR_SERVER_IP:8080',
  // ...
};
```

### Option 2: Use Custom Domain

```javascript
// src/config.js  
const BACKEND_CONFIG = {
  PRODUCTION_URL: 'https://api.yourdomain.com',
  // ...
};
```

## 🔍 Monitoring & Maintenance

### Check Service Status

```bash
# Check if service is running
systemctl status blofin-proxy

# View logs
cd /opt/blofin-proxy
docker-compose logs -f

# Restart if needed
systemctl restart blofin-proxy
```

### Monitor Resource Usage

```bash
# Check system resources
htop

# Check Docker stats
docker stats

# Check disk space
df -h
```

### Updates

```bash
# Update your code
cd /opt/blofin-proxy
git pull  # if using git
docker-compose down
docker-compose build
docker-compose up -d
```

## 🔒 Security Best Practices

### Firewall

```bash
# Only allow necessary ports
ufw allow ssh
ufw allow 80    # HTTP
ufw allow 443   # HTTPS
ufw enable
```

### Auto-Updates

```bash
# Enable automatic security updates
apt install unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades
```

### Monitoring

```bash
# Set up log rotation
cat > /etc/logrotate.d/blofin-proxy << 'EOF'
/opt/blofin-proxy/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    notifempty
    create 644 root root
}
EOF
```

## 💰 Cost Breakdown

### Monthly Costs:
- **VPS**: $3.50-6/month
- **Domain** (optional): $1/month
- **Total**: $4.50-7/month

### vs Current Netlify:
- **1 active user**: Save $29-32/month
- **10 active users**: Save $353-356/month  
- **100 active users**: Save $3,593-3,596/month

## 🎉 Success!

After setup, you'll have:

- ✅ **Ultra-low cost**: $3.50-6/month total
- ✅ **High performance**: No cold starts, always fast
- ✅ **Full control**: Customize, monitor, scale as needed
- ✅ **Same security**: Credentials still client-side only
- ✅ **Global reach**: Choose server location closest to users

Your Blofin API proxy is now running on dedicated hardware for a fraction of the serverless cost! 🚀
