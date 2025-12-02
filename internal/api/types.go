package api

import "time"

// LoginResponse represents the response from login endpoint
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	ValidFrom *time.Time `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to"`
	VpnIP     string     `json:"vpn_ip"`
}

// UserListResponse represents paginated list of users
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// GroupWithNetworks represents a group with its networks
type GroupWithNetworks struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Networks    []Network `json:"networks"`
}

// GroupsResponse represents groups API response
type GroupsResponse struct {
	Groups []GroupWithNetworks `json:"groups"`
}

// Network represents a network/route
type Network struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CIDR        string `json:"cidr"`
	Description string `json:"description"`
}

// VpnSession represents a VPN session
type VpnSession struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	VpnIP          string     `json:"vpn_ip"`
	ClientIP       string     `json:"client_ip"`
	ConnectedAt    time.Time  `json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at"`
	BytesReceived  int64      `json:"bytes_received"`
	BytesSent      int64      `json:"bytes_sent"`
}

// VpnAuthRequest represents VPN authentication request
type VpnAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// VpnAuthResponse represents VPN authentication response
type VpnAuthResponse struct {
	Valid   bool         `json:"valid"`
	User    UserResponse `json:"user"`
	Message string       `json:"message,omitempty"`
}

// CreateSessionRequest represents request to create VPN session
type CreateSessionRequest struct {
	UserID      string `json:"user_id"`
	VpnIP       string `json:"vpn_ip"`
	ClientIP    string `json:"client_ip"`
	ConnectedAt string `json:"connected_at"`
}

// DisconnectSessionRequest represents request to disconnect VPN session
type DisconnectSessionRequest struct {
	DisconnectedAt   string `json:"disconnected_at"`
	BytesReceived    int64  `json:"bytes_received"`
	BytesSent        int64  `json:"bytes_sent"`
	DisconnectReason string `json:"disconnect_reason"`
}

// ErrorResponse represents API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// VpnUsersResponse represents VPN users list response
type VpnUsersResponse struct {
	Users []UserResponse `json:"users"`
}

// RoutesResponse represents routes list response
type RoutesResponse struct {
	Routes []Network `json:"routes"`
}
