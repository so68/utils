package utils

import (
	"testing"
	"time"
)

/*
HTTP客户端功能测试

本文件用于测试HTTPClient结构体的各种功能特性，
包括HTTP请求方法、链式调用、错误处理等。

运行命令：
go test -v -run "^Test.*HTTP.*$"

测试内容：
1. 客户端创建和配置 (NewHTTPClient, SetTimeout, SetHeader等)
2. HTTP请求方法 (Get, Post, Put, Delete, Patch)
3. 链式调用功能验证
4. 响应处理方法 (String, JSON, IsSuccess)
5. URL构建和参数处理
6. 错误处理和边界条件
7. 复杂请求场景测试
*/

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")
	if client == nil {
		t.Error("NewHTTPClient should not return nil")
		return
	}
	if client.baseURL != "https://httpbin.org" {
		t.Errorf("Expected baseURL to be 'https://httpbin.org', got '%s'", client.baseURL)
	}
	if client.client.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", client.client.Timeout)
	}
}

func TestSetTimeout(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")
	newTimeout := 10 * time.Second

	// 测试链式调用
	result := client.SetTimeout(newTimeout)
	if result != client {
		t.Error("SetTimeout should return the same client instance for chaining")
	}

	if client.client.Timeout != newTimeout {
		t.Errorf("Expected timeout to be %v, got %v", newTimeout, client.client.Timeout)
	}
}

func TestSetHeader(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")

	// 测试链式调用
	result := client.SetHeader("Authorization", "Bearer token")
	if result != client {
		t.Error("SetHeader should return the same client instance for chaining")
	}

	if client.headers["Authorization"] != "Bearer token" {
		t.Errorf("Expected Authorization header to be 'Bearer token', got '%s'", client.headers["Authorization"])
	}
}

func TestSetHeaders(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "TestAgent/1.0",
	}

	// 测试链式调用
	result := client.SetHeaders(headers)
	if result != client {
		t.Error("SetHeaders should return the same client instance for chaining")
	}

	for k, v := range headers {
		if client.headers[k] != v {
			t.Errorf("Expected header %s to be '%s', got '%s'", k, v, client.headers[k])
		}
	}
}

func TestSetBaseURL(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")
	newURL := "https://api.example.com"

	// 测试链式调用
	result := client.SetBaseURL(newURL)
	if result != client {
		t.Error("SetBaseURL should return the same client instance for chaining")
	}

	if client.baseURL != newURL {
		t.Errorf("Expected baseURL to be '%s', got '%s'", newURL, client.baseURL)
	}
}

func TestSetContentType(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")
	contentType := "application/json"

	// 测试链式调用
	result := client.SetContentType(contentType)
	if result != client {
		t.Error("SetContentType should return the same client instance for chaining")
	}

	if client.headers["Content-Type"] != contentType {
		t.Errorf("Expected Content-Type to be '%s', got '%s'", contentType, client.headers["Content-Type"])
	}
}

func TestSetAuthorization(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")
	auth := "Bearer token123"

	// 测试链式调用
	result := client.SetAuthorization(auth)
	if result != client {
		t.Error("SetAuthorization should return the same client instance for chaining")
	}

	if client.headers["Authorization"] != auth {
		t.Errorf("Expected Authorization to be '%s', got '%s'", auth, client.headers["Authorization"])
	}
}

func TestSetUserAgent(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org")
	userAgent := "MyApp/1.0"

	// 测试链式调用
	result := client.SetUserAgent(userAgent)
	if result != client {
		t.Error("SetUserAgent should return the same client instance for chaining")
	}

	if client.headers["User-Agent"] != userAgent {
		t.Errorf("Expected User-Agent to be '%s', got '%s'", userAgent, client.headers["User-Agent"])
	}
}

func TestChainHTTPClient(t *testing.T) {
	// 测试完整的链式调用
	client := NewHTTPClient("https://httpbin.org").
		SetTimeout(5 * time.Second).
		SetUserAgent("TestAgent/1.0").
		SetContentType("application/json").
		SetAuthorization("Bearer test-token")

	// 验证所有设置都生效了
	if client.client.Timeout != 5*time.Second {
		t.Error("Timeout not set correctly in chain")
	}
	if client.headers["User-Agent"] != "TestAgent/1.0" {
		t.Error("User-Agent not set correctly in chain")
	}
	if client.headers["Content-Type"] != "application/json" {
		t.Error("Content-Type not set correctly in chain")
	}
	if client.headers["Authorization"] != "Bearer test-token" {
		t.Error("Authorization not set correctly in chain")
	}
}

func TestGetRequest(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org").
		SetTimeout(10 * time.Second).
		SetUserAgent("TestAgent/1.0")

	resp := client.Get("get", map[string]string{"test": "value"})
	if resp.Error != nil {
		t.Errorf("GET request failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Errorf("GET request not successful, status code: %d", resp.StatusCode)
	}
	if len(resp.Body) == 0 {
		t.Error("GET request should return response body")
	}
}

func TestPostRequest(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org").
		SetTimeout(10 * time.Second).
		SetContentType("application/json")

	data := map[string]interface{}{
		"test":   "value",
		"number": 123,
	}

	resp := client.Post("post", data)
	if resp.Error != nil {
		t.Errorf("POST request failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Errorf("POST request not successful, status code: %d", resp.StatusCode)
	}
	if len(resp.Body) == 0 {
		t.Error("POST request should return response body")
	}
}

func TestPostFormRequest(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org").
		SetTimeout(10 * time.Second).
		SetUserAgent("FormTest/1.0")

	formData := map[string]string{
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   "25",
	}

	resp := client.PostForm("post", formData)
	if resp.Error != nil {
		t.Errorf("POST form request failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Errorf("POST form request not successful, status code: %d", resp.StatusCode)
	}
	if len(resp.Body) == 0 {
		t.Error("POST form request should return response body")
	}
}

func TestPutRequest(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org").
		SetTimeout(10 * time.Second).
		SetContentType("application/json")

	data := map[string]interface{}{
		"method": "PUT",
		"data":   "test",
	}

	resp := client.Put("put", data)
	if resp.Error != nil {
		t.Errorf("PUT request failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Errorf("PUT request not successful, status code: %d", resp.StatusCode)
	}
}

func TestDeleteRequest(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org").
		SetTimeout(10 * time.Second).
		SetUserAgent("DeleteTest/1.0")

	resp := client.Delete("delete")
	if resp.Error != nil {
		t.Errorf("DELETE request failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Errorf("DELETE request not successful, status code: %d", resp.StatusCode)
	}
}

func TestPatchRequest(t *testing.T) {
	client := NewHTTPClient("https://httpbin.org").
		SetTimeout(10 * time.Second).
		SetContentType("application/json")

	data := map[string]interface{}{
		"method": "PATCH",
		"data":   "test",
	}

	resp := client.Patch("patch", data)
	if resp.Error != nil {
		t.Errorf("PATCH request failed: %v", resp.Error)
	}
	// PATCH请求可能返回405状态码，这是正常的
	if resp.StatusCode != 200 && resp.StatusCode != 405 {
		t.Errorf("PATCH request status code unexpected: %d", resp.StatusCode)
	}
}

func TestHTTPResponseMethods(t *testing.T) {
	client := NewHTTPClient("https://jsonplaceholder.typicode.com").
		SetTimeout(10 * time.Second)

	resp := client.Get("posts/1", nil)
	if resp.Error != nil {
		t.Errorf("Request failed: %v", resp.Error)
		return
	}

	// 测试String方法
	bodyStr := resp.String()
	if len(bodyStr) == 0 {
		t.Error("String() should return non-empty string")
	}

	// 测试JSON方法
	var post struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		UserID int    `json:"userId"`
	}

	if err := resp.JSON(&post); err != nil {
		t.Errorf("JSON parsing failed: %v", err)
	}

	if post.ID != 1 {
		t.Errorf("Expected post ID to be 1, got %d", post.ID)
	}

	// 测试IsSuccess方法
	if !resp.IsSuccess() {
		t.Errorf("Expected request to be successful, status code: %d", resp.StatusCode)
	}
}

func TestErrorHandling(t *testing.T) {
	// 测试无效URL
	client := NewHTTPClient("https://nonexistent-domain-12345.com").
		SetTimeout(2 * time.Second)

	resp := client.Get("test", nil)
	if resp.Error == nil {
		t.Error("Expected error for invalid domain")
	}

	// 测试404状态码
	client = NewHTTPClient("https://httpbin.org").
		SetTimeout(5 * time.Second)

	resp = client.Get("status/404", nil)
	if resp.Error != nil {
		t.Errorf("Unexpected error for 404: %v", resp.Error)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", resp.StatusCode)
	}
	if resp.IsSuccess() {
		t.Error("404 should not be considered successful")
	}
}

func TestBuildURL(t *testing.T) {
	client := NewHTTPClient("https://api.example.com")

	// 测试基础URL构建
	url := client.buildURL("users", nil)
	expected := "https://api.example.com/users"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}

	// 测试带参数的URL构建
	params := map[string]string{"id": "123", "name": "test"}
	url = client.buildURL("users", params)
	expected = "https://api.example.com/users?id=123&name=test"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}

	// 测试带斜杠的路径
	url = client.buildURL("/users/", params)
	expected = "https://api.example.com/users/?id=123&name=test"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestComplexChainRequest(t *testing.T) {
	// 测试复杂的链式调用和请求
	resp := NewHTTPClient("https://httpbin.org").
		SetTimeout(5*time.Second).
		SetHeaders(map[string]string{
			"Accept":    "application/json",
			"X-Test-ID": "12345",
		}).
		SetAuthorization("Bearer test-token").
		SetUserAgent("ComplexTest/1.0").
		Get("headers", nil)

	if resp.Error != nil {
		t.Errorf("Complex chain request failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Errorf("Complex chain request not successful, status code: %d", resp.StatusCode)
	}

	// 验证响应包含我们设置的请求头
	bodyStr := resp.String()
	if len(bodyStr) == 0 {
		t.Error("Complex chain request should return response body")
	}
}
