# OpenVPN Sample Configurations

Sample configurations for OpenVPN server and client integrated with OpenVPN Manager.

## Server Configurations

| File | Protocol | Port | Description |
|------|----------|------|-------------|
| `server.conf` | UDP | 1194 | **Default** - best performance |
| `server-tcp.conf` | TCP | 993 | For restrictive networks (see below) |

## Client Configurations

| File | Protocol | Port | Description |
|------|----------|------|-------------|
| `client.ovpn` | UDP | 1194 | **Default** - best performance |
| `client-tcp.ovpn` | TCP | 993 | For restrictive networks (see below) |

## Why TCP on Port 993?

Some networks (corporate, hotel, airport) block VPN traffic:
- Standard VPN ports (1194, 1723) are blocked
- UDP protocol may be completely blocked

**Solution:** Use TCP on port 993 (IMAPS - secure email)
- Email ports are rarely blocked
- TCP traffic looks like normal HTTPS/email
- Alternative ports: 443 (HTTPS), 587 (SMTP)

**Trade-off:** TCP has higher latency than UDP due to double error correction.

## Key Points

### No CCD (Client Config Dir) Needed

With OpenVPN Manager integration, you **don't need** `client-config-dir`:

| Feature | Traditional CCD | OpenVPN Manager |
|---------|-----------------|-----------------|
| Static IP | File per user in ccd/ | User's `vpn_ip` field in database |
| Routes | File per user in ccd/ | Groups → Networks in database |
| Management | Manual file editing | Web UI |

The `openvpn-connect` script dynamically generates client config from API.

### No Client Certificates Required

Clients authenticate with **username/password only**:

- No need to generate certificates per user
- No certificate distribution headaches
- Revocation handled by disabling user in OpenVPN Manager
- Simplifies client deployment

**Required on client:**
- CA certificate (to verify server)
- TLS-Auth key (optional but recommended)

### Redirect Gateway (Full Tunnel)

To route **all traffic** through VPN (client appears as VPN server IP):

1. In OpenVPN Manager, create network `0.0.0.0/0`
2. Assign this network to user's group
3. `openvpn-connect` will automatically push `redirect-gateway def1`

**Split tunnel** (only specific networks via VPN):
- Assign only specific networks (e.g., `192.168.1.0/24`) to group
- Client's internet traffic goes directly, only matched traffic via VPN

## Server Setup

### Directory Structure

```
/etc/openvpn/
├── pki/
│   ├── ca.crt           # CA certificate
│   ├── server.crt       # Server certificate
│   ├── server.key       # Server private key
│   ├── dh.pem           # Diffie-Hellman parameters
│   ├── crl.pem          # Certificate Revocation List
│   └── ta.key           # TLS-Auth key
├── client/
│   └── config.yaml      # OpenVPN Client config
└── server/
    ├── server.conf      # UDP config (default)
    └── server-tcp.conf  # TCP config (optional)
```

### Required Files

Generate PKI (if not existing):
```bash
# Install easy-rsa
apt install easy-rsa  # Debian/Ubuntu
yum install easy-rsa  # RHEL/CentOS

# Initialize PKI
cd /etc/openvpn
easyrsa init-pki
easyrsa build-ca nopass
easyrsa gen-dh
easyrsa build-server-full server nopass
openvpn --genkey secret pki/ta.key

# Create empty CRL
easyrsa gen-crl
```

### Firewall (NFTables)

```nft
# Allow OpenVPN
udp dport 1194 accept  # UDP (default)
tcp dport 993 accept   # TCP (optional)

# NAT for VPN clients
table inet nat {
    chain postrouting {
        type nat hook postrouting priority 100;
        ip saddr 10.90.90.0/24 masquerade
    }
}

# Enable IP forwarding
sysctl -w net.ipv4.ip_forward=1
```

### Enable and Start

```bash
# Copy config
cp server.conf /etc/openvpn/server/

# Enable and start
systemctl enable openvpn-server@server
systemctl start openvpn-server@server

# For TCP (optional, run both)
cp server-tcp.conf /etc/openvpn/server/
systemctl enable openvpn-server@server-tcp
systemctl start openvpn-server@server-tcp
```

## Client Setup

### 1. Get Files from Server

```bash
# On server
cat /etc/openvpn/pki/ca.crt
cat /etc/openvpn/pki/ta.key
```

### 2. Edit Client Config

1. Copy `client.ovpn` (or `client-tcp.ovpn` for restrictive networks)
2. Replace `vpn.example.com` with your server address
3. Paste CA certificate into `<ca>` section
4. Paste TLS-Auth key into `<tls-auth>` section

### 3. Connect

Import into OpenVPN client and connect with your OpenVPN Manager credentials.

## Encryption

Default settings (server and client must match):
- **Cipher**: AES-256-GCM
- **Auth**: SHA256
- **TLS-Auth**: Enabled

## Troubleshooting

**Client can't connect:**
```bash
# Check server is running
systemctl status openvpn-server@server

# Check port is listening
ss -ulnp | grep 1194  # UDP
ss -tlnp | grep 993   # TCP

# Check firewall
nft list ruleset | grep -E "1194|993"
```

**Authentication fails:**
```bash
# Test login script manually
echo -e "username\npassword" > /tmp/auth.txt
/usr/local/bin/openvpn-login /tmp/auth.txt
echo $?  # 0 = success
```

**Routes not pushed:**
```bash
# Check connect script
common_name=testuser /usr/local/bin/openvpn-connect /tmp/test.conf
cat /tmp/test.conf
```
