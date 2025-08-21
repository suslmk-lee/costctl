package nhncloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cost-collect/pkg/config"
	"cost-collect/pkg/storage"
)

type Client struct {
	config    *config.NHNCloudConfig
	token     string
	tokenExp  time.Time
	httpClient *http.Client
}

type AuthRequest struct {
	Auth NHNAuth `json:"auth"`
}

type NHNAuth struct {
	TenantID            string              `json:"tenantId"`
	PasswordCredentials PasswordCredentials `json:"passwordCredentials"`
}

type PasswordCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Access Access `json:"access"`
}

type Access struct {
	Token          AccessToken      `json:"token"`
	ServiceCatalog []ServiceCatalog `json:"serviceCatalog"`
}

type AccessToken struct {
	ID      string `json:"id"`
	Expires string `json:"expires"`
	Tenant  Tenant `json:"tenant"`
}

type Tenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ServiceCatalog struct {
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Region      string `json:"region"`
	PublicURL   string `json:"publicURL"`
	InternalURL string `json:"internalURL"`
}

type NovaServer struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Status         string                 `json:"status"`
	Created        string                 `json:"created"`
	Updated        string                 `json:"updated"`
	Flavor         map[string]interface{} `json:"flavor"`
	OSExtSTSPower  int                    `json:"OS-EXT-STS:power_state"`
}

type NovaServersResponse struct {
	Servers []NovaServer `json:"servers"`
}

func NewClient(cfg *config.NHNCloudConfig) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) authenticate() error {
	authReq := AuthRequest{
		Auth: NHNAuth{
			TenantID: c.config.TenantID,
			PasswordCredentials: PasswordCredentials{
				Username: c.config.Username,
				Password: c.config.Password,
			},
		},
	}

	jsonData, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("인증 요청 마샬링 실패: %w", err)
	}

	identityURL := c.config.IdentityURL
	if identityURL == "" {
		identityURL = "https://api-identity-infrastructure.nhncloudservice.com"
	}

	req, err := http.NewRequest("POST", identityURL+"/v2.0/tokens", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("인증 요청 생성 실패: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("인증 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("인증 실패 (상태코드: %d): %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("인증 응답 파싱 실패: %w", err)
	}

	c.token = authResp.Access.Token.ID

	// 토큰 만료 시간 파싱
	if authResp.Access.Token.Expires != "" {
		if expTime, err := time.Parse(time.RFC3339, authResp.Access.Token.Expires); err == nil {
			c.tokenExp = expTime
		} else {
			// 파싱 실패 시 1시간으로 설정
			c.tokenExp = time.Now().Add(1 * time.Hour)
		}
	} else {
		c.tokenExp = time.Now().Add(1 * time.Hour)
	}

	return nil
}

func (c *Client) isTokenValid() bool {
	return c.token != "" && time.Now().Before(c.tokenExp)
}

func (c *Client) ensureAuthenticated() error {
	if !c.isTokenValid() {
		return c.authenticate()
	}
	return nil
}

func (c *Client) GetInstances() ([]*storage.InstanceState, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, fmt.Errorf("인증 실패: %w", err)
	}

	// NHN Cloud의 Nova API 엔드포인트
	computeURL := c.config.ComputeURL
	if computeURL == "" {
		computeURL = "https://kr1-api-instance-infrastructure.nhncloudservice.com"
	}
	url := fmt.Sprintf("%s/v2/%s/servers/detail", computeURL, c.config.TenantID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("인스턴스 목록 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("인스턴스 목록 조회 실패 (상태코드: %d): %s", resp.StatusCode, string(body))
	}

	var novaResp NovaServersResponse
	if err := json.NewDecoder(resp.Body).Decode(&novaResp); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	instances := make([]*storage.InstanceState, 0, len(novaResp.Servers))
	now := time.Now()

	for _, server := range novaResp.Servers {
		createdAt, err := time.Parse(time.RFC3339, server.Created)
		if err != nil {
			createdAt = now
		}

		// Updated 시간 파싱 (상태 변경 시간)
		updatedAt, err := time.Parse(time.RFC3339, server.Updated)
		if err != nil {
			updatedAt = now
		}

		// Flavor ID 추출
		flavorID := ""
		if flavor, ok := server.Flavor["id"].(string); ok {
			flavorID = flavor
		}

		// Power State 추출
		powerState := server.OSExtSTSPower
		if powerState == 0 {
			powerState = 1 // 기본값
		}

		instance := &storage.InstanceState{
			ID:                server.ID,
			Name:              server.Name,
			FlavorID:          flavorID,
			CurrentStatus:     server.Status,
			CurrentPowerState: powerState,
			CreatedAt:         createdAt,
			LastUpdated:       updatedAt, // API의 updated 시간 사용
		}

		instances = append(instances, instance)
	}

	return instances, nil
}