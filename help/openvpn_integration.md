# OpenVPN Server Integration Guide

Complete guide for setting up OpenVPN server with OpenVPN Manager integration. This guide covers installation, PKI setup, firewall configuration, and client deployment.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Server Installation](#server-installation)
- [PKI Setup (Certificates)](#pki-setup-certificates)
- [OpenVPN Server Configuration](#openvpn-server-configuration)
- [Firewall Configuration](#firewall-configuration)
- [OpenVPN Client Integration](#openvpn-client-integration)
- [OpenVPN Manager Configuration](#openvpn-manager-configuration)
- [Client Configuration for Users](#client-configuration-for-users)
- [Testing](#testing)
- [Maintenance](#maintenance)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before starting, ensure you have:

1. **OpenVPN Manager** running and accessible (see [openvpn-mng](https://github.com/tldr-it-stepankutaj/openvpn-mng))
2. A server with:
   - Public IP address
   - Root/sudo access
   - Ports 1194/UDP (or your chosen port) open in cloud firewall
3. DNS record pointing to your server (e.g., `vpn.example.com`)

---

## Server Installation

### Debian/Ubuntu

```bash
# Update system
apt update && apt upgrade -y

# Install OpenVPN and Easy-RSA
apt install -y openvpn easy-rsa

# Install additional tools
apt install -y curl jq

# Create directories
mkdir -p /etc/openvpn/{pki,server,client}
mkdir -p /var/log/openvpn
mkdir -p /var/run/openvpn

# Create openvpn user (if not exists)
useradd -r -s /usr/sbin/nologin openvpn 2>/dev/null || true
chown openvpn:openvpn /var/run/openvpn
```

### RHEL/CentOS/Rocky/Alma

```bash
# Enable EPEL repository
dnf install -y epel-release

# Update system
dnf update -y

# Install OpenVPN and Easy-RSA
dnf install -y openvpn easy-rsa

# Install additional tools
dnf install -y curl jq

# Create directories
mkdir -p /etc/openvpn/{pki,server,client}
mkdir -p /var/log/openvpn
mkdir -p /var/run/openvpn

# Create openvpn user (if not exists)
useradd -r -s /sbin/nologin openvpn 2>/dev/null || true
chown openvpn:openvpn /var/run/openvpn

# Disable SELinux (or configure policy)
setenforce 0
sed -i 's/SELINUX=enforcing/SELINUX=permissive/' /etc/selinux/config
```

---

## PKI Setup (Certificates)

### Initialize PKI

```bash
# Navigate to OpenVPN directory
cd /etc/openvpn

# Initialize PKI structure
easyrsa init-pki

# Set PKI variables (optional - edit pki/vars)
cat > pki/vars << 'EOF'
set_var EASYRSA_REQ_COUNTRY    "CZ"
set_var EASYRSA_REQ_PROVINCE   "Czech Republic"
set_var EASYRSA_REQ_CITY       "Prague"
set_var EASYRSA_REQ_ORG        "My Company"
set_var EASYRSA_REQ_EMAIL      "admin@example.com"
set_var EASYRSA_REQ_OU         "IT"
set_var EASYRSA_KEY_SIZE       4096
set_var EASYRSA_ALGO           rsa
set_var EASYRSA_CA_EXPIRE      3650
set_var EASYRSA_CERT_EXPIRE    3650
EOF
```

### Generate CA Certificate

```bash
cd /etc/openvpn

# Build CA (without password for automated operation)
easyrsa build-ca nopass

# Output: pki/ca.crt, pki/private/ca.key
```

### Generate Server Certificate

```bash
cd /etc/openvpn

# Generate server keypair and certificate
easyrsa build-server-full server nopass

# Output: pki/issued/server.crt, pki/private/server.key
```

### Generate Diffie-Hellman Parameters

```bash
cd /etc/openvpn

# Generate DH parameters (takes a few minutes)
easyrsa gen-dh

# Output: pki/dh.pem
```

### Generate TLS-Auth Key

```bash
cd /etc/openvpn

# Generate TLS authentication key
openvpn --genkey secret pki/ta.key

# Output: pki/ta.key
```

### Generate Certificate Revocation List

```bash
cd /etc/openvpn

# Generate initial CRL
easyrsa gen-crl

# Output: pki/crl.pem
```

### Set Permissions

```bash
# Secure private keys
chmod 600 /etc/openvpn/pki/private/*
chmod 600 /etc/openvpn/pki/ta.key

# Allow OpenVPN to read certificates
chmod 644 /etc/openvpn/pki/ca.crt
chmod 644 /etc/openvpn/pki/issued/server.crt
chmod 644 /etc/openvpn/pki/dh.pem
chmod 644 /etc/openvpn/pki/crl.pem
```

### Verify PKI Setup

```bash
# List all generated files
ls -la /etc/openvpn/pki/

# Verify CA certificate
openssl x509 -in /etc/openvpn/pki/ca.crt -noout -subject -dates

# Verify server certificate
openssl x509 -in /etc/openvpn/pki/issued/server.crt -noout -subject -dates

# Verify certificate chain
openssl verify -CAfile /etc/openvpn/pki/ca.crt /etc/openvpn/pki/issued/server.crt
```

---

## OpenVPN Server Configuration

### Create Server Configuration

```bash
cat > /etc/openvpn/server/server.conf << 'EOF'
# OpenVPN Server Configuration (UDP)
# Integrated with OpenVPN Manager API

### NETWORK ###
mode server
tls-server
local 0.0.0.0
port 1194
proto udp
dev tun
topology subnet

# VPN subnet - adjust as needed
server 10.8.0.0 255.255.255.0
push "topology subnet"

### CERTIFICATES ###
ca /etc/openvpn/pki/ca.crt
cert /etc/openvpn/pki/issued/server.crt
key /etc/openvpn/pki/private/server.key
dh /etc/openvpn/pki/dh.pem
crl-verify /etc/openvpn/pki/crl.pem
tls-auth /etc/openvpn/pki/ta.key 0

### ENCRYPTION ###
cipher AES-256-GCM
data-ciphers AES-256-GCM:AES-128-GCM:CHACHA20-POLY1305
auth SHA256

### CONNECTION ###
keepalive 10 120

### PRIVILEGES ###
user openvpn
group openvpn
persist-key
persist-tun

### LOGGING ###
status /var/log/openvpn/status.log
log-append /var/log/openvpn/openvpn.log
verb 3

### MANAGEMENT ###
management localhost 7505

### OPENVPN MANAGER INTEGRATION ###
username-as-common-name
auth-user-pass-verify /usr/local/bin/openvpn-login via-file
client-connect /usr/local/bin/openvpn-connect
client-disconnect /usr/local/bin/openvpn-disconnect
script-security 2
reneg-sec 0
EOF
```

### Enable IP Forwarding

```bash
# Enable immediately
sysctl -w net.ipv4.ip_forward=1

# Make persistent
echo "net.ipv4.ip_forward = 1" > /etc/sysctl.d/99-openvpn.conf
sysctl -p /etc/sysctl.d/99-openvpn.conf
```

---

## Firewall Configuration

### NFTables (Recommended for modern systems)

#### Debian/Ubuntu

```bash
# Install nftables
apt install -y nftables

# Enable service
systemctl enable nftables
systemctl start nftables
```

#### RHEL/CentOS

```bash
# Install nftables
dnf install -y nftables

# Disable firewalld if running
systemctl stop firewalld
systemctl disable firewalld

# Enable nftables
systemctl enable nftables
systemctl start nftables
```

#### NFTables Configuration

```bash
# Create nftables configuration
cat > /etc/nftables.conf << 'EOF'
#!/usr/sbin/nft -f

flush ruleset

table inet filter {
    chain input {
        type filter hook input priority 0; policy drop;

        # Allow established connections
        ct state established,related accept

        # Allow loopback
        iif lo accept

        # Allow SSH
        tcp dport 22 accept

        # Allow OpenVPN (UDP)
        udp dport 1194 accept

        # Allow ICMP
        ip protocol icmp accept
        ip6 nexthdr icmpv6 accept
    }

    chain forward {
        type filter hook forward priority 0; policy drop;

        # Allow established connections
        ct state established,related accept

        # VPN user rules (auto-generated by openvpn-firewall)
        include "/etc/nftables.d/vpn-users.nft"
    }

    chain output {
        type filter hook output priority 0; policy accept;
    }
}

table inet nat {
    chain postrouting {
        type nat hook postrouting priority 100;

        # Masquerade VPN traffic
        ip saddr 10.8.0.0/24 masquerade
    }
}
EOF

# Create directory for VPN rules
mkdir -p /etc/nftables.d

# Create empty VPN rules file
touch /etc/nftables.d/vpn-users.nft

# Apply configuration
nft -f /etc/nftables.conf

# Verify
nft list ruleset
```

### IPTables (Legacy systems)

#### Installation

```bash
# Debian/Ubuntu
apt install -y iptables iptables-persistent

# RHEL/CentOS
dnf install -y iptables-services
systemctl enable iptables
systemctl start iptables
```

#### IPTables Configuration

```bash
# Flush existing rules
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X

# Default policies
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Allow SSH
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# Allow OpenVPN
iptables -A INPUT -p udp --dport 1194 -j ACCEPT

# Allow ICMP
iptables -A INPUT -p icmp -j ACCEPT

# Create VPN_USERS chain
iptables -N VPN_USERS
iptables -A FORWARD -i tun0 -j VPN_USERS

# NAT for VPN clients
iptables -t nat -A POSTROUTING -s 10.8.0.0/24 -j MASQUERADE

# Save rules
# Debian/Ubuntu
netfilter-persistent save

# RHEL/CentOS
service iptables save
```

#### Create IPTables Rules Directory

```bash
mkdir -p /etc/iptables.d
touch /etc/iptables.d/vpn-users.rules
```

---

## OpenVPN Client Integration

### Install OpenVPN Client Binaries

```bash
# Download latest release (adjust URL for your platform)
# Option 1: Build from source
cd /tmp
git clone https://github.com/tldr-it-stepankutaj/openvpn-client.git
cd openvpn-client
make build-linux

# Install binaries
cp build/linux-amd64/openvpn-* /usr/local/bin/
chmod 755 /usr/local/bin/openvpn-*

# Verify installation
ls -la /usr/local/bin/openvpn-*
```

### Configure OpenVPN Client

```bash
# Create configuration directory
mkdir -p /etc/openvpn/client

# Create configuration file
cat > /etc/openvpn/client/config.yaml << 'EOF'
api:
  # OpenVPN Manager API URL
  base_url: "http://127.0.0.1:8080"

  # API token (get from OpenVPN Manager config)
  token: "your-vpn-token-here"

  # Request timeout
  timeout: 10s

openvpn:
  # Directory for session files
  session_dir: "/var/run/openvpn"

firewall:
  # Firewall type: "nftables" or "iptables"
  type: "nftables"

  nftables:
    rules_file: "/etc/nftables.d/vpn-users.nft"
    reload_command: "/usr/sbin/nft -f /etc/nftables.conf"

  # Use this section if using iptables
  # iptables:
  #   chain_name: "VPN_USERS"
  #   rules_file: "/etc/iptables.d/vpn-users.rules"
  #   reload_command: "iptables-restore -n < /etc/iptables.d/vpn-users.rules"
EOF

# Secure configuration
chmod 600 /etc/openvpn/client/config.yaml
```

### Generate API Token

```bash
# Generate secure token
openssl rand -hex 32

# Output example: a1b2c3d4e5f6...
# Copy this token and add to:
# 1. /etc/openvpn/client/config.yaml (token field)
# 2. OpenVPN Manager config.yaml (api.vpn_token field)
```

### Setup Firewall Cron Job

```bash
# Create cron job to update firewall rules
cat > /etc/cron.d/openvpn-firewall << 'EOF'
# Update VPN firewall rules every 5 minutes
*/5 * * * * root /usr/local/bin/openvpn-firewall -c /etc/openvpn/client/config.yaml >> /var/log/openvpn/firewall.log 2>&1
EOF

# Run initial firewall update
/usr/local/bin/openvpn-firewall -c /etc/openvpn/client/config.yaml
```

---

## OpenVPN Manager Configuration

### Add VPN Token to OpenVPN Manager

Edit OpenVPN Manager's `config.yaml`:

```yaml
api:
  enabled: true
  # Use the same token generated above
  vpn_token: "your-vpn-token-here"
```

Restart OpenVPN Manager to apply changes.

### Create Network for Full Tunnel

To route all client traffic through VPN:

1. Go to OpenVPN Manager web UI
2. Navigate to **Networks**
3. Create new network:
   - Name: `Internet (Full Tunnel)`
   - CIDR: `0.0.0.0/0`
   - Description: `Route all traffic through VPN`
4. Assign this network to groups that need full tunnel

### Create Users and Groups

1. **Create Groups** (e.g., "Administrators", "Developers", "Users")
2. **Assign Networks** to groups based on access requirements
3. **Create Users** and assign them to groups
4. **Set VPN IP** for each user (optional - for static IP assignment)

---

## Client Configuration for Users

### Create Universal Client Configuration

This is the configuration file you'll distribute to all users:

```bash
cat > /etc/openvpn/client-dist/client.ovpn << 'EOF'
# VPN Client Configuration
# Company: My Company
# Server: vpn.example.com

client
dev tun
proto udp
remote vpn.example.com 1194
nobind
persist-key
persist-tun

<ca>
EOF

# Append CA certificate
cat /etc/openvpn/pki/ca.crt >> /etc/openvpn/client-dist/client.ovpn

cat >> /etc/openvpn/client-dist/client.ovpn << 'EOF'
</ca>

<tls-auth>
EOF

# Append TLS auth key
cat /etc/openvpn/pki/ta.key >> /etc/openvpn/client-dist/client.ovpn

cat >> /etc/openvpn/client-dist/client.ovpn << 'EOF'
</tls-auth>
key-direction 1

remote-cert-tls server
cipher AES-256-GCM
auth-nocache
auth-user-pass
resolv-retry infinite
verb 3
EOF
```

### Create Client Distribution Directory

```bash
mkdir -p /etc/openvpn/client-dist
chmod 755 /etc/openvpn/client-dist
```

### Generate Client Config Script

Create a script to generate client configs:

```bash
cat > /usr/local/bin/generate-client-config << 'EOF'
#!/bin/bash

# Configuration
OUTPUT_DIR="/etc/openvpn/client-dist"
SERVER_ADDR="vpn.example.com"
SERVER_PORT="1194"
CA_CERT="/etc/openvpn/pki/ca.crt"
TA_KEY="/etc/openvpn/pki/ta.key"

# Generate config
cat > "${OUTPUT_DIR}/client.ovpn" << CONF
# VPN Client Configuration
# Generated: $(date)
# Server: ${SERVER_ADDR}
#
# Instructions:
# 1. Import this file into your OpenVPN client
# 2. Connect using your username and password
# 3. Routes will be automatically assigned based on your group

client
dev tun
proto udp
remote ${SERVER_ADDR} ${SERVER_PORT}
nobind
persist-key
persist-tun

<ca>
$(cat ${CA_CERT})
</ca>

<tls-auth>
$(cat ${TA_KEY})
</tls-auth>
key-direction 1

remote-cert-tls server
cipher AES-256-GCM
auth-nocache
auth-user-pass
resolv-retry infinite
verb 3
CONF

echo "Client configuration generated: ${OUTPUT_DIR}/client.ovpn"
EOF

chmod +x /usr/local/bin/generate-client-config
```

### Generate and Distribute

```bash
# Generate client configuration
/usr/local/bin/generate-client-config

# Verify
cat /etc/openvpn/client-dist/client.ovpn

# Distribute to users via:
# - Email (encrypted)
# - Secure file share
# - Internal portal
# - OpenVPN Manager download (if implemented)
```

---

## Testing

### Start OpenVPN Server

```bash
# Enable and start service
systemctl enable openvpn-server@server
systemctl start openvpn-server@server

# Check status
systemctl status openvpn-server@server

# View logs
journalctl -u openvpn-server@server -f
```

### Test Authentication

```bash
# Create test credentials
echo -e "testuser\ntestpassword" > /tmp/auth-test.txt

# Test login script
/usr/local/bin/openvpn-login -c /etc/openvpn/client/config.yaml /tmp/auth-test.txt
echo "Exit code: $?"

# Cleanup
rm /tmp/auth-test.txt
```

### Test Connection Script

```bash
# Simulate client connect
export common_name="testuser"
export trusted_ip="1.2.3.4"
export trusted_port="12345"
export ifconfig_pool_remote_ip="10.8.0.10"

/usr/local/bin/openvpn-connect -c /etc/openvpn/client/config.yaml /tmp/client-config.txt

# Check generated config
cat /tmp/client-config.txt

# Cleanup
rm /tmp/client-config.txt
unset common_name trusted_ip trusted_port ifconfig_pool_remote_ip
```

### Test Firewall Rules

```bash
# Generate rules (dry run)
/usr/local/bin/openvpn-firewall -c /etc/openvpn/client/config.yaml -n

# Apply rules
/usr/local/bin/openvpn-firewall -c /etc/openvpn/client/config.yaml

# Verify nftables
nft list ruleset | grep -A 10 "vpn-users"

# Or verify iptables
iptables -L VPN_USERS -n -v
```

### Test Client Connection

```bash
# On a client machine
openvpn --config client.ovpn

# Or import into OpenVPN Connect app and connect
```

### Verify Connection

```bash
# On server - check status
cat /var/log/openvpn/status.log

# Check connected clients
cat /var/log/openvpn/status.log | grep "CLIENT_LIST"

# On client - verify tunnel
ip addr show tun0

# Test connectivity
ping 10.8.0.1  # VPN gateway
```

---

## Maintenance

### Certificate Renewal

```bash
# Check certificate expiration
openssl x509 -in /etc/openvpn/pki/issued/server.crt -noout -enddate

# Renew server certificate (before expiration)
cd /etc/openvpn
easyrsa renew server nopass

# Restart OpenVPN
systemctl restart openvpn-server@server
```

### Update CRL

```bash
# Regenerate CRL (do this periodically or after revoking certs)
cd /etc/openvpn
easyrsa gen-crl

# Restart OpenVPN to pick up new CRL
systemctl restart openvpn-server@server
```

### Log Rotation

```bash
# Create logrotate configuration
cat > /etc/logrotate.d/openvpn << 'EOF'
/var/log/openvpn/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 640 root openvpn
    postrotate
        systemctl reload openvpn-server@server 2>/dev/null || true
    endscript
}
EOF
```

### Backup

```bash
# Backup PKI and configuration
tar -czf /backup/openvpn-$(date +%Y%m%d).tar.gz \
    /etc/openvpn/pki \
    /etc/openvpn/server \
    /etc/openvpn/client

# Store securely off-server
```

---

## Troubleshooting

### OpenVPN Server Won't Start

```bash
# Check configuration syntax
openvpn --config /etc/openvpn/server/server.conf --verb 4

# Check permissions
ls -la /etc/openvpn/pki/

# Check if port is in use
ss -ulnp | grep 1194

# Check logs
journalctl -u openvpn-server@server --no-pager -n 50
```

### Authentication Fails

```bash
# Test API connectivity
curl -H "X-VPN-Token: your-token" http://127.0.0.1:8080/api/v1/vpn-auth/users

# Check OpenVPN Manager is running
systemctl status openvpn-manager

# Test login manually
echo -e "user\npass" > /tmp/auth.txt
/usr/local/bin/openvpn-login -c /etc/openvpn/client/config.yaml /tmp/auth.txt
echo $?
```

### Client Can't Connect

```bash
# Check server is listening
ss -ulnp | grep 1194

# Check firewall
nft list ruleset | grep 1194
# or
iptables -L -n | grep 1194

# Test from client
nc -vzu vpn.example.com 1194
```

### Routes Not Working

```bash
# Check connect script output
tail -f /var/log/openvpn/openvpn.log

# Test connect script manually
common_name=testuser /usr/local/bin/openvpn-connect /tmp/test.conf
cat /tmp/test.conf

# Check user's groups and networks in OpenVPN Manager
curl -H "X-VPN-Token: your-token" \
    http://127.0.0.1:8080/api/v1/vpn-auth/users/by-username/testuser
```

### Firewall Rules Not Applied

```bash
# Check firewall script
/usr/local/bin/openvpn-firewall -c /etc/openvpn/client/config.yaml -n

# Check rules file
cat /etc/nftables.d/vpn-users.nft

# Manually reload
nft -f /etc/nftables.conf
```

### Performance Issues

```bash
# Check MTU
ping -M do -s 1472 10.8.0.1

# Add to server config if needed:
# mssfix 1400
# fragment 1400

# Check CPU usage (encryption)
top -p $(pgrep openvpn)

# Consider using AES-NI
grep -o aes /proc/cpuinfo | head -1
```

---

## Quick Reference

### File Locations

| File | Purpose |
|------|---------|
| `/etc/openvpn/server/server.conf` | Server configuration |
| `/etc/openvpn/client/config.yaml` | Client integration config |
| `/etc/openvpn/pki/` | PKI certificates and keys |
| `/etc/nftables.d/vpn-users.nft` | Auto-generated firewall rules |
| `/var/log/openvpn/` | Log files |
| `/var/run/openvpn/` | Session files |
| `/etc/openvpn/client-dist/client.ovpn` | Client config for distribution |

### Commands

| Command | Purpose |
|---------|---------|
| `systemctl status openvpn-server@server` | Check server status |
| `journalctl -u openvpn-server@server -f` | Follow server logs |
| `openvpn-firewall -n` | Preview firewall rules |
| `openvpn-firewall` | Apply firewall rules |
| `nft list ruleset` | Show nftables rules |
| `/usr/local/bin/generate-client-config` | Generate client config |

### Ports

| Port | Protocol | Service |
|------|----------|---------|
| 1194 | UDP | OpenVPN (default) |
| 993 | TCP | OpenVPN (firewall bypass) |
| 7505 | TCP | OpenVPN management |
| 8080 | TCP | OpenVPN Manager API |
