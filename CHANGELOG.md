# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-02-06

### Added
- `StatusCode` field on `VpnAuthResponse` to track HTTP status codes from the server

### Changed
- `ValidateVpnUser()` now parses server error response body to extract specific error messages (rate limit, account lockout) instead of returning generic "authentication failed"
- **openvpn-login** distinguishes HTTP 429 (rate limited / account locked) from other authentication failures in log output

### Fixed
- Error responses from VPN auth endpoint no longer silently discarded — server-provided messages are now propagated to the caller

## [1.0.0] - 2025-12-02

### Added

#### Core OpenVPN Integration Binaries
- **openvpn-login** - User authentication against OpenVPN Manager API via `auth-user-pass-verify` directive
- **openvpn-connect** - Client connection handler with dynamic IP assignment and route pushing based on user group membership
- **openvpn-disconnect** - Session termination handler with traffic statistics reporting (bytes sent/received)
- **openvpn-firewall** - Firewall rules generator based on user-network assignments from API

#### API Client
- REST API client for OpenVPN Manager integration
- Dual authentication support:
  - API Token authentication (recommended, via `X-VPN-Token` header)
  - Legacy service account authentication (JWT Bearer token)
- Endpoints for user authentication, session management, and route/network retrieval

#### Firewall Support
- Plugin-based architecture for firewall implementations
- **nftables** backend - Modern firewall with `ip saddr` match syntax
- **iptables** backend - Legacy support with custom `VPN_USERS` chain management
- Atomic rule reload with drift detection
- Dynamic rules based on user VPN IP and allowed networks

#### Configuration System
- YAML configuration file support (default: `/etc/openvpn/client/config.yaml`)
- Environment variable overrides (prefix: `OPENVPN_*`)
- CLI argument support with highest priority
- Validation for authentication credentials and firewall type

#### Logging
- Structured JSON logging using Go 1.21+ `log/slog`
- Context helpers for user and session tracking
- Program name tagging for log aggregation
- Configurable log levels

#### Session Management
- Session file storage in `/var/run/openvpn/`
- Session ID, client IP, and port tracking
- Traffic accounting (bytes sent/received) on disconnect

#### Build System
- Makefile with cross-compilation targets (Linux amd64, arm64)
- Version, commit, and build-time injection via ldflags
- GitHub Actions CI/CD pipeline

#### Utilities
- CIDR to netmask conversion for OpenVPN route configuration
- Default route (`0.0.0.0/0`) handling for redirect-gateway scenarios

[1.1.0]: https://github.com/tldr-it-stepankutaj/openvpn-client/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/tldr-it-stepankutaj/openvpn-client/releases/tag/v1.0.0
