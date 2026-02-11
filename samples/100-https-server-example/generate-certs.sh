#!/bin/bash

# Generate self-signed certificates for HTTPS development/testing
# WARNING: Do NOT use these certificates in production!

echo "🔐 Generating self-signed TLS certificates..."
echo ""

# Generate private key
openssl genrsa -out server.key 2048

# Generate certificate signing request (CSR)
openssl req -new -key server.key -out server.csr \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Generate self-signed certificate (valid for 365 days)
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt

# Clean up CSR
rm server.csr

echo ""
echo "✅ Certificates generated successfully!"
echo ""
echo "📁 Files created:"
echo "   - server.crt (Certificate)"
echo "   - server.key (Private Key)"
echo ""
echo "⚠️  WARNING: These are self-signed certificates for development only."
echo "   For production, use certificates from a trusted CA (e.g., Let's Encrypt)"
echo ""
