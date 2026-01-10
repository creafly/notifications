package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type VerifyResponse struct {
	Valid       bool       `json:"valid"`
	UserID      uuid.UUID  `json:"userId,omitempty"`
	Email       string     `json:"email,omitempty"`
	TenantID    *uuid.UUID `json:"tenantId,omitempty"`
	Claims      []string   `json:"claims,omitempty"`
	IsBlocked   bool       `json:"isBlocked,omitempty"`
	BlockReason *string    `json:"blockReason,omitempty"`
	BlockedAt   *time.Time `json:"blockedAt,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type ValidateTenantResponse struct {
	Valid    bool   `json:"valid"`
	IsMember bool   `json:"isMember"`
	TenantID string `json:"tenantId,omitempty"`
	UserID   string `json:"userId,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Error    string `json:"error,omitempty"`
}

type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type UsersListResponse struct {
	Users []User `json:"users"`
}

type IdentityClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewIdentityClient(baseURL string) *IdentityClient {
	return &IdentityClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *IdentityClient) VerifyToken(ctx context.Context, accessToken string) (*VerifyResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/auth/verify", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call identity service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var verifyResp VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &verifyResp, nil
}

func (c *IdentityClient) ValidateTenantAccess(ctx context.Context, accessToken string, tenantID uuid.UUID) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/validate", c.baseURL, tenantID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call identity service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, nil
	}

	var validateResp ValidateTenantResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return validateResp.Valid && validateResp.IsMember, nil
}

func (c *IdentityClient) GetAllUsers(ctx context.Context, accessToken string) ([]uuid.UUID, error) {
	url := fmt.Sprintf("%s/api/v1/users?limit=10000", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call identity service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("identity service returned status %d", resp.StatusCode)
	}

	var usersResp UsersListResponse
	if err := json.NewDecoder(resp.Body).Decode(&usersResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	userIDs := make([]uuid.UUID, len(usersResp.Users))
	for i, user := range usersResp.Users {
		userIDs[i] = user.ID
	}

	return userIDs, nil
}

func (c *IdentityClient) GetTenantMembers(ctx context.Context, accessToken string, tenantID uuid.UUID) ([]uuid.UUID, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/members?limit=10000", c.baseURL, tenantID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call identity service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("identity service returned status %d", resp.StatusCode)
	}

	var members []User
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	userIDs := make([]uuid.UUID, len(members))
	for i, member := range members {
		userIDs[i] = member.ID
	}

	return userIDs, nil
}
