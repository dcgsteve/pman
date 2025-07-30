# PMAN Installation Guide

This guide covers installation and deployment of the PMAN password manager.

## Quick Installation

### Prerequisites

- Go 1.19+ (for building from source)
- SQLite3 (for backend database)
- Git (for version information during build)

### Option 1: Pre-built Binaries (Recommended)

1. Download the latest release for your platform from the releases page
2. Extract the archive:
   ```bash
   # Linux/macOS
   tar -xzf pman-v1.0.0-linux-amd64.tar.gz
   cd pman-v1.0.0-linux-amd64/
   
   # Windows
   unzip pman-v1.0.0-windows-amd64.zip
   cd pman-v1.0.0-windows-amd64/
   ```

3. Copy binaries to your PATH:
   ```bash
   # Linux/macOS
   sudo cp pman /usr/local/bin/
   sudo cp pman-backend /usr/local/bin/
   
   # Windows (as Administrator)
   copy pman.exe C:\Windows\System32\
   copy pman-backend.exe C:\Windows\System32\
   ```

### Option 2: Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/pman.git
   cd pman
   ```

2. Build using Make (recommended):
   ```bash
   make build
   ```

3. Or use the build script:
   ```bash
   ./build.sh
   ```

4. Install locally:
   ```bash
   make install
   ```

## Backend Deployment

### Environment Variables

The backend requires these environment variables:

```bash
# Required
export PMAN_ENCRYPTION_KEY="your-32-character-encryption-key"
export PMAN_DOMAIN_NAME="your-server.example.com"
export PMAN_DEFAULT_EXPIRE_DAYS="7"

# Optional
export PORT="5000"                    # Default: 5000
export DATABASE_PATH="/path/to/db"    # Default: ./pman.db
```

### Production Deployment Options

#### Option 1: Systemd Service (Linux)

1. Create service file `/etc/systemd/system/pman-backend.service`:
   ```ini
   [Unit]
   Description=PMAN Password Manager Backend
   After=network.target
   
   [Service]
   Type=simple
   User=pman
   Group=pman
   WorkingDirectory=/opt/pman
   Environment=PMAN_ENCRYPTION_KEY=your-32-character-key
   Environment=PMAN_DOMAIN_NAME=your-server.example.com
   Environment=PMAN_DEFAULT_EXPIRE_DAYS=7
   Environment=PORT=5000
   ExecStart=/usr/local/bin/pman-backend
   Restart=always
   RestartSec=5
   
   [Install]
   WantedBy=multi-user.target
   ```

2. Enable and start:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable pman-backend
   sudo systemctl start pman-backend
   ```

#### Option 2: Docker Deployment

Create `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ./build.sh

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

COPY --from=builder /app/build/pman-backend .

EXPOSE 5000
CMD ["./pman-backend"]
```

Build and run:
```bash
docker build -t pman-backend .
docker run -d \
  -p 5000:5000 \
  -e PMAN_ENCRYPTION_KEY="your-key" \
  -e PMAN_DOMAIN_NAME="your-domain" \
  -e PMAN_DEFAULT_EXPIRE_DAYS="7" \
  -v pman-data:/root \
  --name pman-backend \
  pman-backend
```

#### Option 3: Reverse Proxy Setup (Nginx)

```nginx
server {
    listen 80;
    server_name your-pman-server.com;

    location / {
        proxy_pass http://localhost:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# SSL configuration with Let's Encrypt
server {
    listen 443 ssl;
    server_name your-pman-server.com;

    ssl_certificate /etc/letsencrypt/live/your-pman-server.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-pman-server.com/privkey.pem;

    location / {
        proxy_pass http://localhost:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Client Configuration

### First Time Setup

1. Install the CLI (see above)
2. Login to your PMAN server:
   ```bash
   pman login -s https://your-pman-server.com -u admin@pman.system -p DefaultPassword
   ```

3. Change the default admin password:
   ```bash
   pman passwd
   ```

4. Set your default group:
   ```bash
   pman setgroup team1
   ```

### Configuration Locations

The CLI stores configuration in:
- **Linux/macOS**: `~/.pman/config.json`
- **Windows**: `%USERPROFILE%\.pman\config.json`

Configuration is encrypted using machine-specific keys for security.

## Security Considerations

### Backend Security

1. **Use strong encryption keys**: Generate a 32-character random key:
   ```bash
   openssl rand -base64 32
   ```

2. **Enable HTTPS**: Always use SSL/TLS in production
3. **Firewall rules**: Restrict access to backend port
4. **Regular updates**: Keep the software updated
5. **Backup database**: Regular SQLite database backups

### Client Security

1. **Secure endpoints**: Only connect to HTTPS endpoints
2. **Token management**: Tokens are automatically encrypted locally
3. **Logout when done**: Use `pman logout` on shared machines

## Troubleshooting

### Common Issues

1. **Connection refused**: Check if backend is running and port is correct
2. **Invalid token**: Re-login with `pman login`
3. **Permission denied**: Check user groups and access rights
4. **Database locked**: Ensure only one backend instance is running

### Debug Information

- Check backend logs: `journalctl -u pman-backend` (systemd)
- Test connectivity: `pman status -s https://your-server.com`
- Verify version: `pman version`

### Log Locations

- **Backend logs**: Check systemd journal or container logs
- **Client errors**: Displayed in terminal
- **Database**: Default location `./pman.db` in backend working directory

## Backup and Recovery

### Database Backup

```bash
# Stop backend
sudo systemctl stop pman-backend

# Backup database
cp /opt/pman/pman.db /backup/pman-$(date +%Y%m%d).db

# Start backend
sudo systemctl start pman-backend
```

### Configuration Backup

```bash
# Backup client config
cp ~/.pman/config.json ~/.pman/config.json.backup
```

## Performance Tuning

### Backend Optimization

1. **Database tuning**: SQLite performs well for most use cases
2. **Resource limits**: Set appropriate memory/CPU limits
3. **Connection pooling**: Built-in Go HTTP server handles this
4. **Monitoring**: Use tools like Prometheus for monitoring

### Scaling Considerations

For high-volume deployments:
1. Consider load balancing multiple backend instances
2. Shared database setup may be needed
3. Token cleanup job scheduling
4. Log rotation and monitoring

## Migration from Other Password Managers

PMAN provides a simple API that can be used to import from other systems. Contact the maintainers for migration scripts for specific password managers.

## Support

- Documentation: See CLAUDE.md for detailed feature documentation
- Issues: Report bugs on the project repository
- Security issues: Report privately to maintainers