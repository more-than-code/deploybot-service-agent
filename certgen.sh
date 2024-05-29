#!/bin/bash

DST_DIR="/etc/nginx/sites-enabled/"

# Function to display usage information
usage() {
    echo "Usage: $0 -d <DOMAIN_GROUP> -f <FILENAME> -l <LISTEN_PORT> -p <PROXY_PORT>"
    exit 1
}

# Parse command-line arguments
while getopts ":d:f:l:p:" opt; do
    case $opt in
        d) DOMAIN_GROUP="$OPTARG";;
        f) FILENAME="$OPTARG";;
        l) LISTEN_PORT="$OPTARG";;
        p) PROXY_PORT="$OPTARG";;
        *) usage;;
    esac
done

# Check if all required arguments are provided
if [[ -z $DOMAIN_GROUP || -z $FILENAME || -z $LISTEN_PORT || -z $PROXY_PORT ]]; then
    usage
fi


# Extract the first domain from the domain group
DOMAIN=$(echo "$DOMAIN_GROUP" | awk -F, '{print $1}')

echo "Domain: $DOMAIN"

# Use certbot to retrieve the certificate information
CERT_INFO=$(certbot certificates 2>/dev/null)

# Ensure CERT_INFO was retrieved successfully
if [ $? -ne 0 ]; then
    echo "Failed to retrieve certificate information."
    exit 1
fi


# Extract the certificate path
CERT_PATH=$(echo "$CERT_INFO" | grep -A 3 "Domains: $DOMAIN" | grep "Certificate Path" | awk '{print $3}')

# Extract the private key path
KEY_PATH=$(echo "$CERT_INFO" | grep -A 3 "Domains: $DOMAIN" | grep "Private Key Path" | awk '{print $4}')

# Print the paths
echo "Certificate Path: $CERT_PATH"
echo "Private Key Path: $KEY_PATH"

# Generate certificate if not exist
if [ ! -f "$CERT_PATH" ] || [ ! -f "$KEY_PATH" ]; then
    echo "Certificate or key file not found, generating..."
    certbot certonly --nginx -d "$DOMAIN_GROUP"

    # Update CERT_INFO after certificate generation
    CERT_INFO=$(certbot certificates 2>/dev/null)

    # Re-extract the certificate and key paths
    CERT_PATH=$(echo "$CERT_INFO" | grep -A 3 "Domains: $DOMAIN" | grep "Certificate Path" | awk '{print $3}' | xargs)
    KEY_PATH=$(echo "$CERT_INFO" | grep -A 3 "Domains: $DOMAIN" | grep "Private Key Path" | awk '{print $4}' | xargs)

    # Check again if the paths were found after generation
    if [ -z "$CERT_PATH" ] || [ -z "$KEY_PATH" ]; then
        echo "Failed to generate or retrieve certificate paths for domain $DOMAIN after generation."
        exit 1
    fi
fi

# Convert the domain group to a space-separated string
DOMAINS=$(echo "$DOMAIN_GROUP" | sed 's/,/ /g')

# Generate Nginx server block configuration
NGINX_CONFIG="server {
    listen $LISTEN_PORT ssl;
    server_name $DOMAINS;

    ssl_certificate $CERT_PATH;
    ssl_certificate_key $KEY_PATH;

    # Optional: Add SSL parameters
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256';

    location / {
        proxy_pass http://127.0.0.1:$PROXY_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}"

# Write the Nginx configuration to a file
CONFIG_FILE="$DST_DIR$FILENAME"
echo "$NGINX_CONFIG" | tee "$CONFIG_FILE" > /dev/null

# Test and reload Nginx
nginx -t && systemctl reload nginx

echo "Nginx configuration for $DOMAIN_GROUP has been created and enabled."