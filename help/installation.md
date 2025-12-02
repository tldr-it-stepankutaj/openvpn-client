# Installation Guide

This guide covers the complete installation and configuration of the OpenVPN Client for integration with OpenVPN Manager API.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Building from Source](#building-from-source)
- [Installation](#installation)
- [Configuration](#configuration)
- [OpenVPN Server Setup](#openvpn-server-setup)
- [Firewall Setup](#firewall-setup)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Server Requirements

- Linux server (RHEL/CentOS/Rocky/Alma 8+ or Debian/Ubuntu 20.04+)
- OpenVPN 2.4+ installed and configured
- OpenVPN Manager API running and accessible
- Root or sudo access

### Build Requirements

- Go 1.23 or later
- Make
- Git

---

## Building from Source

### Clone Repository

```bash
git clone https://github.com/tldr-it-stepankutaj/openvpn-client.git
cd openvpn-client
```

### Build Binaries

```bash
# Build for current platform
make build

# Build for Linux amd64 (for deployment to server)
make build-linux

# Build for Linux arm64
make build-linux-arm64
```

The binaries will be created in the `build/` directory:
- `openvpn-login`
- `openvpn-connect`
- `openvpn-disconnect`
- `openvpn-firewall`

---

## Installation

### 1. Copy Binaries

```bash
# From build machine to target server
scp build/linux-amd64/* root@vpn-server:/usr/local/bin/

# Or install locally
sudo make install
```

### 2. Set Permissions

```bash
sudo chmod 755 /usr/local/bin/openvpn-*
```

### 3. Create Directories

```bash
# Configuration directory
sudo mkdir -p /etc/openvpn/client

# Session files directory
sudo mkdir -p /var/run/openvpn
sudo chown openvpn:openvpn /var/run/openvpn

# Firewall rules directory (nftables)
sudo mkdir -p /etc/nftables.d

# Or for iptables
sudo mkdir -p /etc/iptables.d
```

---

## Configuration

### 1. Create Configuration File

```bash
sudo cp config.example.yaml /etc/openvpn/client/config.yaml
sudo chmod 600 /etc/openvpn/client/config.yaml
```

### 2. Edit Configuration

```bash
sudo vim /etc/openvpn/client/config.yaml
```

#### Minimal Configuration (API Token)

```yaml
api:
  base_url: "http://127.0.0.1:8080"
  token: "your-api-token-from-openvpn-manager"
  timeout: 10s

openvpn:
  session_dir: "/var/run/openvpn"

firewall:
  type: "nftables"
  nftables:
    rules_file: "/etc/nftables.d/vpn-users.nft"
    reload_command: "/usr/sbin/nft -f /etc/sysconfig/nftables.conf"
```

#### Full Configuration (with iptables)

```yaml
api:
  base_url: "http://127.0.0.1:8080"
  token: "your-api-token-from-openvpn-manager"
  timeout: 10s

openvpn:
  session_dir: "/var/run/openvpn"

firewall:
  type: "iptables"
  iptables:
    chain_name: "VPN_USERS"
    rules_file: "/etc/iptables.d/vpn-users.rules"
    reload_command: "iptables-restore -n < /etc/iptables.d/vpn-users.rules"
```

### 3. Configure OpenVPN Manager

Add VPN token to your OpenVPN Manager `config.yaml`:

```yaml
api:
  enabled: true
  vpn_token: "your-api-token-from-openvpn-manager"
```

Generate a secure token:

```bash
openssl rand -hex 32
```

---

## OpenVPN Server Setup

### 1. Update OpenVPN Server Configuration

Add these lines to your OpenVPN server configuration:

```conf
# Use username as common name (required)
username-as-common-name

# Authentication script
auth-user-pass-verify /usr/local/bin/openvpn-login via-file

# Client connect/disconnect scripts
client-connect /usr/local/bin/openvpn-connect
client-disconnect /usr/local/bin/openvpn-disconnect

# Enable script execution
script-security 2

# Disable renegotiation (prevents script re-execution)
reneg-sec 0
```

### 2. Restart OpenVPN

```bash
# Systemd
sudo systemctl restart openvpn@server

# Or for specific instance
sudo systemctl restart openvpn@myserver
```

---

## Firewall Setup

### NFTables Setup

#### 1. Create Empty Rules File

```bash
sudo touch /etc/nftables.d/vpn-users.nft
```

#### 2. Update Main NFTables Configuration

Edit `/etc/sysconfig/nftables.conf` or `/etc/nftables.conf`:

```nft
table inet filter {
    chain forward {
        type filter hook forward priority 0; policy drop;
        ct state established,related accept

        # Include auto-generated VPN user rules
        include "/etc/nftables.d/vpn-users.nft"
    }
}

table inet nat {
    chain postrouting {
        type nat hook postrouting priority 100;
        ip saddr 10.8.0.0/24 masquerade
    }
}
```

#### 3. Initial Rules Generation

```bash
sudo /usr/local/bin/openvpn-firewall
```

### IPTables Setup

#### 1. Create Empty Rules File

```bash
sudo touch /etc/iptables.d/vpn-users.rules
```

#### 2. Update Main IPTables Configuration

```bash
# Add VPN_USERS chain to FORWARD
iptables -N VPN_USERS
iptables -A FORWARD -i tun0 -j VPN_USERS

# Add NAT masquerade
iptables -t nat -A POSTROUTING -s 10.8.0.0/24 -j MASQUERADE
```

#### 3. Initial Rules Generation

```bash
sudo /usr/local/bin/openvpn-firewall
```

### Cron Job for Automatic Updates

```bash
# Create cron job
echo '*/5 * * * * root /usr/local/bin/openvpn-firewall >> /var/log/openvpn-firewall.log 2>&1' | sudo tee /etc/cron.d/openvpn-firewall
```

---

## Verification

### Test API Connectivity

```bash
# Test with API token
curl -H "X-VPN-Token: your-token" http://127.0.0.1:8080/api/v1/vpn-auth/users
```

### Test Authentication

```bash
# Create test auth file
echo -e "testuser\ntestpassword" > /tmp/auth-test.txt

# Run login script
/usr/local/bin/openvpn-login -c /etc/openvpn/client/config.yaml /tmp/auth-test.txt
echo "Exit code: $?"

# Cleanup
rm /tmp/auth-test.txt
```

### Test Firewall Generation

```bash
# Dry run - print rules without applying
/usr/local/bin/openvpn-firewall -c /etc/openvpn/client/config.yaml -n
```

### Check Logs

```bash
# OpenVPN logs
sudo tail -f /var/log/openvpn.log

# Firewall update logs
sudo tail -f /var/log/openvpn-firewall.log

# System journal
sudo journalctl -u openvpn@server -f
```

---

## Troubleshooting

### Common Issues

#### 1. Authentication Fails

**Symptoms:** Users cannot connect, login script exits with code 1

**Check:**
- API connectivity: `curl http://127.0.0.1:8080/api/v1/health`
- API token is correct in both client and server config
- User exists and is active in OpenVPN Manager

#### 2. Routes Not Pushed

**Symptoms:** User connects but cannot reach internal networks

**Check:**
- User has assigned VPN IP in OpenVPN Manager
- User is member of groups with network assignments
- Check connect script output in OpenVPN log

#### 3. Firewall Rules Not Applied

**Symptoms:** Connected users cannot reach allowed networks

**Check:**
- Firewall rules file exists and is readable
- Include statement in main firewall config is correct
- Reload command works: `sudo nft -f /etc/sysconfig/nftables.conf`

#### 4. Session Files Not Created

**Symptoms:** Disconnect statistics not recorded

**Check:**
- Session directory exists: `/var/run/openvpn`
- Directory is writable by OpenVPN user
- Check permissions: `ls -la /var/run/openvpn`

### Debug Mode

Run commands manually with verbose output:

```bash
# Test authentication
common_name=testuser /usr/local/bin/openvpn-login /tmp/auth.txt 2>&1

# Test connect
common_name=testuser trusted_ip=1.2.3.4 ifconfig_pool_remote_ip=10.8.0.10 \
  /usr/local/bin/openvpn-connect /tmp/client.conf 2>&1

# Test firewall
/usr/local/bin/openvpn-firewall -n 2>&1
```

### Log Analysis

All binaries output JSON-formatted logs to stderr:

```json
{"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"user authenticated","program":"openvpn-login","username":"john.doe","user_id":"uuid-here"}
```

Parse logs with `jq`:

```bash
cat /var/log/openvpn.log | grep openvpn-login | jq .
```

---

## SELinux Configuration (RHEL/CentOS)

If SELinux is enabled, you may need to create a policy module:

```bash
# Check for denials
sudo ausearch -m avc -ts recent

# Generate and apply policy
sudo ausearch -c 'openvpn' --raw | audit2allow -M openvpn-client
sudo semodule -i openvpn-client.pp
```

Or set permissive mode for OpenVPN:

```bash
sudo semanage permissive -a openvpn_t
```

---

## Upgrading

### 1. Build New Version

```bash
git pull
make build-linux
```

### 2. Replace Binaries

```bash
sudo systemctl stop openvpn@server
scp build/linux-amd64/* root@vpn-server:/usr/local/bin/
sudo systemctl start openvpn@server
```

### 3. Verify

```bash
/usr/local/bin/openvpn-firewall -n
```
