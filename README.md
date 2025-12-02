# OpenVPN Client

A Go-based client for integrating OpenVPN server with [OpenVPN Manager](https://github.com/tldr-it-stepankutaj/openvpn-mng) API.

> **Important:** This client is designed to work exclusively with [OpenVPN Manager](https://github.com/tldr-it-stepankutaj/openvpn-mng). OpenVPN Manager provides a web-based administration interface for managing VPN users, networks, and groups. This client connects your OpenVPN server to that management system.

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  OpenVPN Server │────▶│  OpenVPN Client │────▶│  OpenVPN Manager│
│                 │     │   (this repo)   │     │   (openvpn-mng) │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │                        │
                               │                        ▼
                               │                 ┌─────────────────┐
                               │                 │    Database     │
                               │                 │ (Users, Groups, │
                               ▼                 │    Networks)    │
                        ┌─────────────────┐      └─────────────────┘
                        │ nftables/iptables│
                        │   (firewall)    │
                        └─────────────────┘
```

**OpenVPN Manager** handles:
- User management (create, edit, delete VPN users)
- Group management (organize users into groups)
- Network management (define internal networks/subnets)
- Access control (assign networks to groups)
- Session monitoring (view active connections)

**OpenVPN Client** (this repository) handles:
- User authentication against OpenVPN Manager API
- Dynamic route assignment based on user's group membership
- VPN session tracking (connect/disconnect events, traffic stats)
- Firewall rules generation based on user-network assignments

## Features

- **User Authentication** - Validates VPN user credentials against the API
- **Client Connect** - Configures client IP, pushes routes based on group membership
- **Client Disconnect** - Records session end and traffic statistics
- **Firewall Rules** - Generates nftables or iptables rules based on user-network assignments
- **Structured Logging** - JSON logging with `log/slog`
- **Flexible Configuration** - CLI arguments, environment variables, and YAML config file

## Quick Start

### Build

```bash
# Build all binaries
make build

# Build for Linux (deployment)
make build-linux
```

### Install

```bash
# Install to /usr/local/bin
sudo make install

# Or manually copy binaries
sudo cp build/openvpn-* /usr/local/bin/
sudo chmod 755 /usr/local/bin/openvpn-*
```

### Configure

```bash
# Create config directory
sudo mkdir -p /etc/openvpn/client

# Copy and edit configuration
sudo cp config.example.yaml /etc/openvpn/client/config.yaml
sudo chmod 600 /etc/openvpn/client/config.yaml
sudo vim /etc/openvpn/client/config.yaml
```

## Binaries

| Binary | Purpose | OpenVPN Directive |
|--------|---------|-------------------|
| `openvpn-login` | User authentication | `auth-user-pass-verify` |
| `openvpn-connect` | Client connection setup | `client-connect` |
| `openvpn-disconnect` | Client disconnection cleanup | `client-disconnect` |
| `openvpn-firewall` | Firewall rules generator | Cron job |

## Configuration

Configuration is loaded with the following priority (highest to lowest):

1. CLI arguments (`-c` / `--config`)
2. Environment variables (`OPENVPN_*`)
3. Configuration file (default: `/etc/openvpn/client/config.yaml`)

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENVPN_CLIENT_CONFIG` | Path to configuration file |
| `OPENVPN_API_BASE_URL` | API base URL |
| `OPENVPN_API_TOKEN` | API token (recommended) |
| `OPENVPN_API_USERNAME` | Service account username (legacy) |
| `OPENVPN_API_PASSWORD` | Service account password (legacy) |
| `OPENVPN_API_TIMEOUT` | API request timeout |
| `OPENVPN_SESSION_DIR` | Session files directory |
| `OPENVPN_FIREWALL_TYPE` | Firewall type (nftables/iptables) |

### Example Configuration

```yaml
api:
  base_url: "http://127.0.0.1:8080"
  token: "your-api-token"
  timeout: 10s

openvpn:
  session_dir: "/var/run/openvpn"

firewall:
  type: "nftables"
  nftables:
    rules_file: "/etc/nftables.d/vpn-users.nft"
    reload_command: "/usr/sbin/nft -f /etc/sysconfig/nftables.conf"
```

## Usage

### Authentication (openvpn-login)

```bash
# Called by OpenVPN with auth file path
openvpn-login [-c /path/to/config.yaml] /tmp/auth.txt
```

### Client Connect (openvpn-connect)

```bash
# Called by OpenVPN with config file path
# Requires environment variables: common_name, trusted_ip, trusted_port, ifconfig_pool_remote_ip
openvpn-connect [-c /path/to/config.yaml] /tmp/client-config.txt
```

### Client Disconnect (openvpn-disconnect)

```bash
# Called by OpenVPN
# Requires environment variables: common_name, bytes_received, bytes_sent
openvpn-disconnect [-c /path/to/config.yaml]
```

### Firewall Rules (openvpn-firewall)

```bash
# Generate and apply firewall rules
openvpn-firewall [-c /path/to/config.yaml]

# Dry run - print rules without applying
openvpn-firewall [-c /path/to/config.yaml] -n
```

## OpenVPN Server Configuration

Add to your OpenVPN server configuration:

```conf
# Authentication via API
username-as-common-name
auth-user-pass-verify /usr/local/bin/openvpn-login via-file
client-connect /usr/local/bin/openvpn-connect
client-disconnect /usr/local/bin/openvpn-disconnect
script-security 2
```

See [samples/openvpn/](samples/openvpn/) for complete examples.

## Firewall Integration

### NFTables

Include the generated rules file in your main nftables configuration:

```nft
chain forward {
    type filter hook forward priority 0; policy drop;
    ct state established,related accept
    include "/etc/nftables.d/vpn-users.nft"
}
```

### IPTables

The generated rules create/flush a custom chain (default: `VPN_USERS`).

### Cron Job

```bash
# Update firewall rules every 5 minutes
*/5 * * * * root /usr/local/bin/openvpn-firewall >> /var/log/openvpn-firewall.log 2>&1
```

## Prerequisites

1. **OpenVPN Manager** must be installed and running - see [openvpn-mng](https://github.com/tldr-it-stepankutaj/openvpn-mng)
2. API token must be configured in OpenVPN Manager
3. Users, groups, and networks must be configured in OpenVPN Manager

## Documentation

| Document | Description |
|----------|-------------|
| [OpenVPN Integration Guide](help/openvpn_integration.md) | **Complete setup guide** - OpenVPN server, PKI, firewall, and integration |
| [Installation Guide](help/installation.md) | Quick installation of this client |
| [Client Integration Guide](help/client.md) | API reference and client details |
| [Sample Configurations](samples/) | OpenVPN server/client configs, firewall examples |

## Related Projects

- [OpenVPN Manager](https://github.com/tldr-it-stepankutaj/openvpn-mng) - Web-based administration for VPN users, networks, and groups

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Tidy dependencies
make tidy
```

## License

[Apache License 2.0](LICENSE)
