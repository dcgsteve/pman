package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/steve/pman/shared/models"
)

type Client struct {
	BaseURL string
	Token   string
	client  *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.BaseURL + "/api/v1" + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.client.Do(req)
}

func (c *Client) Login(email, password string, expireDays int) (*models.LoginResponse, error) {
	loginReq := models.LoginRequest{
		Email:      email,
		Password:   password,
		ExpireDays: expireDays,
	}

	resp, err := c.makeRequest("POST", "/auth/login", loginReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s", string(body))
	}

	var loginResp models.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &loginResp, nil
}

func (c *Client) CheckHealth() error {
	resp, err := c.makeRequest("GET", "/health", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server unhealthy: %s", string(body))
	}

	return nil
}

func (c *Client) CreatePassword(path, value, group string) error {
	passwordReq := models.PasswordRequest{
		Path:  path,
		Value: value,
	}

	endpoint := fmt.Sprintf("/passwords?group=%s", group)
	resp, err := c.makeRequest("POST", endpoint, passwordReq)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create password failed: %s", string(body))
	}

	return nil
}

func (c *Client) GetPassword(path, group string) (string, error) {
	endpoint := fmt.Sprintf("/passwords/%s/%s", group, path)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get password failed: %s", string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return result["value"], nil
}

func (c *Client) ListPasswords(group, pathPrefix string) ([]string, error) {
	endpoint := fmt.Sprintf("/passwords/%s", group)
	if pathPrefix != "" {
		endpoint += fmt.Sprintf("?prefix=%s", pathPrefix)
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list passwords failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	paths, ok := result["paths"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var stringPaths []string
	for _, path := range paths {
		if pathStr, ok := path.(string); ok {
			stringPaths = append(stringPaths, pathStr)
		}
	}

	return stringPaths, nil
}

func (c *Client) UpdatePassword(path, value, group string) error {
	passwordReq := models.PasswordRequest{
		Path:  path,
		Value: value,
	}

	endpoint := fmt.Sprintf("/passwords/%s/%s", group, path)
	resp, err := c.makeRequest("PUT", endpoint, passwordReq)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update password failed: %s", string(body))
	}

	return nil
}

func (c *Client) DeletePassword(path, group string) error {
	endpoint := fmt.Sprintf("/passwords/%s/%s", group, path)
	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete password failed: %s", string(body))
	}

	return nil
}

func (c *Client) DeletePasswordRecursive(path, group string) (int, error) {
	endpoint := fmt.Sprintf("/passwords/%s/%s?recursive=true", group, path)
	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("recursive delete failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %v", err)
	}

	count, ok := result["count"].(float64)
	if !ok {
		return 1, nil // Default to 1 if count not provided
	}

	return int(count), nil
}

func (c *Client) GetPasswordInfo(path, group string) (*models.PasswordInfo, error) {
	endpoint := fmt.Sprintf("/passwords/%s/%s/info", group, path)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get password info failed: %s", string(body))
	}

	var passwordInfo models.PasswordInfo
	if err := json.NewDecoder(resp.Body).Decode(&passwordInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &passwordInfo, nil
}

// User Management Methods (Admin Only)

func (c *Client) CreateUser(email, role, groups string) (string, error) {
	userReq := models.UserRequest{
		Email:  email,
		Role:   role,
		Groups: groups,
	}

	resp, err := c.makeRequest("POST", "/admin/users", userReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create user failed: %s", string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return result["password"], nil
}

func (c *Client) ListUsers() ([]models.User, error) {
	resp, err := c.makeRequest("GET", "/admin/users", nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list users failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	usersData, ok := result["users"].([]interface{})
	if !ok {
		return []models.User{}, nil
	}

	var users []models.User
	for _, userData := range usersData {
		userBytes, _ := json.Marshal(userData)
		var user models.User
		if err := json.Unmarshal(userBytes, &user); err == nil {
			users = append(users, user)
		}
	}

	return users, nil
}

func (c *Client) UpdateUser(email, role, groups string) error {
	userReq := models.UserRequest{
		Email:  email,
		Role:   role,
		Groups: groups,
	}

	endpoint := fmt.Sprintf("/admin/users/%s", email)
	resp, err := c.makeRequest("PUT", endpoint, userReq)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update user failed: %s", string(body))
	}

	return nil
}

func (c *Client) DeleteUser(email string) error {
	endpoint := fmt.Sprintf("/admin/users/%s", email)
	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete user failed: %s", string(body))
	}

	return nil
}

func (c *Client) EnableUser(email string) error {
	endpoint := fmt.Sprintf("/admin/users/%s/enable", email)
	resp, err := c.makeRequest("POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("enable user failed: %s", string(body))
	}

	return nil
}

func (c *Client) DisableUser(email string) error {
	endpoint := fmt.Sprintf("/admin/users/%s/disable", email)
	resp, err := c.makeRequest("POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("disable user failed: %s", string(body))
	}

	return nil
}

func (c *Client) ChangePassword(currentPassword, newPassword string) error {
	req := map[string]string{
		"current_password": currentPassword,
		"new_password":     newPassword,
	}

	resp, err := c.makeRequest("POST", "/auth/passwd", req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("change password failed: %s", string(body))
	}

	return nil
}

func (c *Client) AdminChangePassword(email, newPassword string) error {
	req := map[string]string{
		"new_password": newPassword,
	}

	endpoint := fmt.Sprintf("/admin/users/%s/passwd", email)
	resp, err := c.makeRequest("POST", endpoint, req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("admin change password failed: %s", string(body))
	}

	return nil
}
