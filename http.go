package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient HTTP客户端结构体
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// HTTPResponse HTTP响应结构体
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Error      error
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// SetTimeout 设置请求超时时间（链式调用）
func (c *HTTPClient) SetTimeout(timeout time.Duration) *HTTPClient {
	c.client.Timeout = timeout
	return c
}

// SetHeader 设置请求头（链式调用）
func (c *HTTPClient) SetHeader(key, value string) *HTTPClient {
	c.headers[key] = value
	return c
}

// SetHeaders 批量设置请求头（链式调用）
func (c *HTTPClient) SetHeaders(headers map[string]string) *HTTPClient {
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

// SetBaseURL 设置基础URL（链式调用）
func (c *HTTPClient) SetBaseURL(baseURL string) *HTTPClient {
	c.baseURL = baseURL
	return c
}

// SetContentType 设置Content-Type请求头（链式调用）
func (c *HTTPClient) SetContentType(contentType string) *HTTPClient {
	c.headers["Content-Type"] = contentType
	return c
}

// SetAuthorization 设置Authorization请求头（链式调用）
func (c *HTTPClient) SetAuthorization(auth string) *HTTPClient {
	c.headers["Authorization"] = auth
	return c
}

// SetUserAgent 设置User-Agent请求头（链式调用）
func (c *HTTPClient) SetUserAgent(userAgent string) *HTTPClient {
	c.headers["User-Agent"] = userAgent
	return c
}

// Get 发送GET请求
func (c *HTTPClient) Get(path string, params map[string]string) *HTTPResponse {
	return c.request("GET", path, params, nil)
}

// Post 发送POST请求
func (c *HTTPClient) Post(path string, data interface{}) *HTTPResponse {
	return c.request("POST", path, nil, data)
}

// PostForm 发送表单POST请求
func (c *HTTPClient) PostForm(path string, formData map[string]string) *HTTPResponse {
	return c.requestForm("POST", path, formData)
}

// Put 发送PUT请求
func (c *HTTPClient) Put(path string, data interface{}) *HTTPResponse {
	return c.request("PUT", path, nil, data)
}

// Delete 发送DELETE请求
func (c *HTTPClient) Delete(path string) *HTTPResponse {
	return c.request("DELETE", path, nil, nil)
}

// Patch 发送PATCH请求
func (c *HTTPClient) Patch(path string, data interface{}) *HTTPResponse {
	return c.request("PATCH", path, nil, data)
}

// request 通用请求方法
func (c *HTTPClient) request(method, path string, params map[string]string, data interface{}) *HTTPResponse {
	// 构建完整URL
	fullURL := c.buildURL(path, params)

	// 准备请求体
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return &HTTPResponse{Error: fmt.Errorf("json marshal error: %w", err)}
		}
		body = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return &HTTPResponse{Error: fmt.Errorf("create request error: %w", err)}
	}

	// 设置请求头
	c.setHeaders(req)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return &HTTPResponse{Error: fmt.Errorf("request error: %w", err)}
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{Error: fmt.Errorf("read response error: %w", err)}
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}
}

// requestForm 发送表单请求
func (c *HTTPClient) requestForm(method, path string, formData map[string]string) *HTTPResponse {
	// 构建完整URL
	fullURL := c.buildURL(path, nil)

	// 准备表单数据
	values := url.Values{}
	for k, v := range formData {
		values.Set(k, v)
	}
	body := strings.NewReader(values.Encode())

	// 创建请求
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return &HTTPResponse{Error: fmt.Errorf("create request error: %w", err)}
	}

	// 设置请求头
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return &HTTPResponse{Error: fmt.Errorf("request error: %w", err)}
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{Error: fmt.Errorf("read response error: %w", err)}
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}
}

// buildURL 构建完整URL
func (c *HTTPClient) buildURL(path string, params map[string]string) string {
	fullURL := c.baseURL
	if !strings.HasSuffix(fullURL, "/") {
		fullURL += "/"
	}
	path = strings.TrimPrefix(path, "/")
	fullURL += path

	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		fullURL += "?" + values.Encode()
	}

	return fullURL
}

// setHeaders 设置请求头
func (c *HTTPClient) setHeaders(req *http.Request) {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
}

// String 返回响应体字符串
func (r *HTTPResponse) String() string {
	return string(r.Body)
}

// JSON 解析响应体为JSON
func (r *HTTPResponse) JSON(v interface{}) error {
	if r.Error != nil {
		return r.Error
	}
	return json.Unmarshal(r.Body, v)
}

// IsSuccess 判断请求是否成功
func (r *HTTPResponse) IsSuccess() bool {
	return r.Error == nil && r.StatusCode >= 200 && r.StatusCode < 300
}
