package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
)

const (
	headerContentType = "Content-Type"
	headerVPNToken    = "X-VPN-Token"
	headerAuth        = "Authorization"
	contentTypeJSON   = "application/json"
)

// Client represents the API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string // JWT token (from service account login)
	apiToken   string // Static API token (from config)
}

// NewClient creates a new API client
func NewClient(cfg *config.APIConfig) *Client {
	c := &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}

	if cfg.UseToken() {
		c.apiToken = cfg.Token
	}

	return c
}

// Authenticate gets a JWT token using service account credentials (legacy)
func (c *Client) Authenticate(ctx context.Context, username, password string) error {
	body := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/auth/login", body, false)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.token = loginResp.Token
	return nil
}

// ValidateVpnUser validates VPN user credentials
func (c *Client) ValidateVpnUser(ctx context.Context, username, password string) (*VpnAuthResponse, error) {
	body := VpnAuthRequest{
		Username: username,
		Password: password,
	}

	// Use VPN-specific endpoint if using API token
	endpoint := "/api/v1/vpn-auth/authenticate"
	if c.apiToken == "" {
		// Fallback to regular login for legacy service account
		endpoint = "/api/v1/auth/login"
	}

	resp, err := c.doRequest(ctx, http.MethodPost, endpoint, body, c.apiToken != "")
	if err != nil {
		return nil, fmt.Errorf("validate user request failed: %w", err)
	}
	defer resp.Body.Close()

	if c.apiToken != "" {
		// VPN auth endpoint returns VpnAuthResponse
		if resp.StatusCode != http.StatusOK {
			return &VpnAuthResponse{Valid: false, Message: "authentication failed"}, nil
		}

		var authResp VpnAuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
			return nil, fmt.Errorf("failed to decode auth response: %w", err)
		}
		return &authResp, nil
	}

	// Legacy login endpoint
	if resp.StatusCode != http.StatusOK {
		return &VpnAuthResponse{Valid: false, Message: "authentication failed"}, nil
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}
	return &VpnAuthResponse{Valid: true, User: loginResp.User}, nil
}

// GetUserByUsername finds a user by username
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*UserResponse, error) {
	// Use VPN-specific endpoint if using API token
	if c.apiToken != "" {
		resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/vpn-auth/users/by-username/"+url.PathEscape(username), nil, true)
		if err != nil {
			return nil, fmt.Errorf("get user request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, c.parseError(resp)
		}

		var user UserResponse
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, fmt.Errorf("failed to decode user response: %w", err)
		}
		return &user, nil
	}

	// Legacy: search users
	params := url.Values{}
	params.Set("search", username)
	params.Set("page_size", "100")

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/users?"+params.Encode(), nil, true)
	if err != nil {
		return nil, fmt.Errorf("search users request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var listResp UserListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode users list: %w", err)
	}

	for _, user := range listResp.Users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found: %s", username)
}

// GetUserRoutes gets user's allowed networks (routes)
func (c *Client) GetUserRoutes(ctx context.Context, userID string) ([]Network, error) {
	// Use VPN-specific endpoint if using API token
	endpoint := "/api/v1/vpn-auth/users/" + userID + "/routes"
	if c.apiToken == "" {
		endpoint = "/api/v1/users/" + userID + "/groups"
	}

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil, true)
	if err != nil {
		return nil, fmt.Errorf("get routes request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	if c.apiToken != "" {
		// VPN endpoint returns routes wrapped in object
		var routesResp RoutesResponse
		if err := json.NewDecoder(resp.Body).Decode(&routesResp); err != nil {
			return nil, fmt.Errorf("failed to decode routes: %w", err)
		}
		return routesResp.Routes, nil
	}

	// Legacy: parse groups response
	var groupsResp GroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&groupsResp); err != nil {
		return nil, fmt.Errorf("failed to decode groups: %w", err)
	}

	// Flatten networks from all groups
	networkMap := make(map[string]Network)
	for _, group := range groupsResp.Groups {
		for _, network := range group.Networks {
			networkMap[network.CIDR] = network
		}
	}

	networks := make([]Network, 0, len(networkMap))
	for _, n := range networkMap {
		networks = append(networks, n)
	}
	return networks, nil
}

// GetAllActiveUsers gets all active users for firewall rules
func (c *Client) GetAllActiveUsers(ctx context.Context) ([]UserResponse, error) {
	// Use VPN-specific endpoint if using API token
	if c.apiToken != "" {
		resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/vpn-auth/users", nil, true)
		if err != nil {
			return nil, fmt.Errorf("get users request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, c.parseError(resp)
		}

		var usersResp VpnUsersResponse
		if err := json.NewDecoder(resp.Body).Decode(&usersResp); err != nil {
			return nil, fmt.Errorf("failed to decode users: %w", err)
		}
		return usersResp.Users, nil
	}

	// Legacy: paginated user list
	var allUsers []UserResponse
	page := 1
	pageSize := 100

	for {
		params := url.Values{}
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
		params.Set("is_active", "true")

		resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/users?"+params.Encode(), nil, true)
		if err != nil {
			return nil, fmt.Errorf("get users page %d failed: %w", page, err)
		}

		var listResp UserListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode users page %d: %w", page, err)
		}
		resp.Body.Close()

		allUsers = append(allUsers, listResp.Users...)

		if page >= listResp.TotalPages {
			break
		}
		page++
	}

	return allUsers, nil
}

// CreateSession creates a new VPN session
func (c *Client) CreateSession(ctx context.Context, userID, vpnIP, clientIP string) (*VpnSession, error) {
	body := CreateSessionRequest{
		UserID:      userID,
		VpnIP:       vpnIP,
		ClientIP:    clientIP,
		ConnectedAt: time.Now().UTC().Format(time.RFC3339),
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/vpn-auth/sessions", body, true)
	if err != nil {
		return nil, fmt.Errorf("create session request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var session VpnSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}

	return &session, nil
}

// DisconnectSession ends a VPN session
func (c *Client) DisconnectSession(ctx context.Context, sessionID string, bytesReceived, bytesSent int64) error {
	body := DisconnectSessionRequest{
		DisconnectedAt:   time.Now().UTC().Format(time.RFC3339),
		BytesReceived:    bytesReceived,
		BytesSent:        bytesSent,
		DisconnectReason: "USER_REQUEST",
	}

	resp, err := c.doRequest(ctx, http.MethodPut, "/api/v1/vpn-auth/sessions/"+sessionID+"/disconnect", body, true)
	if err != nil {
		return fmt.Errorf("disconnect session request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, auth bool) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(headerContentType, contentTypeJSON)

	if auth {
		if c.apiToken != "" {
			req.Header.Set(headerVPNToken, c.apiToken)
		} else if c.token != "" {
			req.Header.Set(headerAuth, "Bearer "+c.token)
		}
	}

	return c.httpClient.Do(req)
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
		return fmt.Errorf("%s: %s", errResp.Error, errResp.Message)
	}

	return fmt.Errorf("API error: %s (status %d)", string(body), resp.StatusCode)
}
