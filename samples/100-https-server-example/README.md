# HTTPS Server Example

This example demonstrates how to use HTTPS with N.O.V.A. server agents:
- ServerAgent
- CrewServerAgent
- GatewayServerAgent

## Prerequisites

- Go 1.21 or higher
- OpenSSL (for generating self-signed certificates)

## Quick Start

### 1. Generate Self-Signed Certificates

For development/testing purposes, generate self-signed certificates:

```bash
./generate-certs.sh
```

This will create:
- `server.crt` - Certificate file
- `server.key` - Private key file

### 2. Run the Example

```bash
go run main.go
```

The server will start on `https://localhost:8443`

### 3. Test the HTTPS Server

```bash
# Using curl (accept self-signed certificate)
curl -k https://localhost:8443/health

# Or specify the certificate
curl --cacert server.crt https://localhost:8443/health
```

## Usage Options

### Option 1: Using Certificate Files (Recommended)

```go
agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(8443),
    server.WithTLSCertFromFile("server.crt", "server.key"),
)
```

### Option 2: Using Certificate Data in Memory

```go
certData, _ := os.ReadFile("server.crt")
keyData, _ := os.ReadFile("server.key")

agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(8443),
    server.WithTLSCert(certData, keyData),
)
```

## Production Deployment

For production, use certificates from a trusted Certificate Authority (CA) like:
- Let's Encrypt (free)
- DigiCert
- Sectigo

### Using Let's Encrypt

```bash
# Install certbot
sudo apt-get install certbot

# Generate certificate
sudo certbot certonly --standalone -d yourdomain.com

# Certificates will be in:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/yourdomain.com/privkey.pem
```

Then use:

```go
agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(443),
    server.WithTLSCertFromFile(
        "/etc/letsencrypt/live/yourdomain.com/fullchain.pem",
        "/etc/letsencrypt/live/yourdomain.com/privkey.pem",
    ),
)
```

## Security Notes

⚠️ **Important**:
- Never commit certificate files to version control
- Use environment variables or secure vaults for production credentials
- Self-signed certificates should only be used for development/testing
- Regularly renew certificates before expiration
- Use strong key sizes (2048-bit RSA minimum, 4096-bit recommended)

## Backward Compatibility

HTTPS is **optional**. If you don't provide TLS certificates, the agents will run in HTTP mode (backward compatible):

```go
// This still works - runs on HTTP
agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(8080),
)
```
