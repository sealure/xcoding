// Package helpers provides shared HTTP utilities and test helpers.

package helpers

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/carlmjohnson/requests"
)

// ---- 方法与状态码常量 ----
const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"

	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusConflict            = 409
	StatusInternalServerError = 500
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
	StatusGatewayTimeout      = 504
)

// ---- HTTP 请求助手 ----
// DoRequest 统一的 HTTP 请求助手（禁用重定向，默认 JSON/UA）
func DoRequest(baseURL, method, endpoint string, body any, headers map[string]string) (int, []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var buf bytes.Buffer
	status := 0

	rb := requests.URL(baseURL + endpoint).Method(method)
	rb.Client(&http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }})
	if body != nil {
		rb.BodyJSON(body)
	}
	for k, v := range headers {
		rb.Header(k, v)
	}
	if headers == nil || headers["Accept"] == "" {
		rb.Header("Accept", "application/json")
	}
	if headers == nil || headers["User-Agent"] == "" {
		rb.Header("User-Agent", "xcoding-e2e-go")
	}
	rb.ToBytesBuffer(&buf)
	rb.AddValidator(func(resp *http.Response) error { status = resp.StatusCode; return nil })
	rb.CheckStatus(StatusOK, StatusCreated, StatusNoContent, StatusBadRequest, StatusUnauthorized, StatusForbidden, StatusNotFound, StatusMethodNotAllowed, StatusConflict, StatusInternalServerError, StatusBadGateway, StatusServiceUnavailable, StatusGatewayTimeout)
    if err := rb.Fetch(ctx); err != nil {
        if buf.Len() == 0 {
            buf.WriteString(err.Error())
        }
        return status, buf.Bytes()
    }
    return status, buf.Bytes()
}

// DoRequestWithHeaders HTTP 请求助手（返回状态码、响应体与响应头）
func DoRequestWithHeaders(baseURL, method, endpoint string, body any, headers map[string]string) (int, []byte, map[string]string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var buf bytes.Buffer
	status := 0
	outHeaders := map[string]string{}

	rb := requests.URL(baseURL + endpoint).Method(method)
	rb.Client(&http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }})
	if body != nil { rb.BodyJSON(body) }
	for k, v := range headers { rb.Header(k, v) }
	if headers == nil || headers["Accept"] == "" { rb.Header("Accept", "application/json") }
	if headers == nil || headers["User-Agent"] == "" { rb.Header("User-Agent", "xcoding-e2e-go") }

	rb.ToBytesBuffer(&buf)
	rb.AddValidator(func(resp *http.Response) error {
		status = resp.StatusCode
		for k, vals := range resp.Header {
			if len(vals) == 0 { continue }
			v := vals[0]
			kl := strings.ToLower(k)
			switch kl {
			case "x-user-id":
				outHeaders["X-User-ID"] = v
			case "x-username":
				outHeaders["X-Username"] = v
			case "x-user-role":
				outHeaders["X-User-Role"] = v
			case "x-scopes":
				outHeaders["X-Scopes"] = v
			default:
				outHeaders[k] = v
			}
		}
		return nil
	})
	rb.CheckStatus(StatusOK, StatusCreated, StatusNoContent, StatusBadRequest, StatusUnauthorized, StatusForbidden, StatusNotFound, StatusMethodNotAllowed, StatusConflict, StatusInternalServerError, StatusBadGateway, StatusServiceUnavailable, StatusGatewayTimeout)
    if err := rb.Fetch(ctx); err != nil {
        if buf.Len() == 0 {
            buf.WriteString(err.Error())
        }
    }
    return status, buf.Bytes(), outHeaders
}

// ---- 路由探测与网关可达 ----
// RouteExists 探测 GET 路由存在且返回 JSON
func RouteExists(baseURL, endpoint string) bool { return RouteExistsWithMethod(baseURL, MethodGet, endpoint) }

// RouteExistsWithMethod 探测指定方法的路由存在且返回 JSON
func RouteExistsWithMethod(baseURL, method, endpoint string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var buf bytes.Buffer
	status := 0
	contentType := ""
	_ = requests.URL(baseURL + endpoint).
		Client(&http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}).
		Method(method).
		Header("Accept", "application/json").
		Header("User-Agent", "xcoding-e2e-go").
		ToBytesBuffer(&buf).
		AddValidator(func(resp *http.Response) error { status = resp.StatusCode; contentType = resp.Header.Get("Content-Type"); return nil }).
		CheckStatus(StatusOK, StatusCreated, StatusNoContent, StatusBadRequest, StatusUnauthorized, StatusForbidden, StatusNotFound, StatusMethodNotAllowed).
		Fetch(ctx)
	return status != StatusNotFound && strings.Contains(strings.ToLower(contentType), "application/json")
}

// PingGateway 探测网关可达
func PingGateway(baseURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status := 0
	_ = requests.URL(baseURL).
		Client(&http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}).
		Method(MethodGet).
		Header("Accept", "application/json").
		Header("User-Agent", "xcoding-e2e-go").
		AddValidator(func(resp *http.Response) error { status = resp.StatusCode; return nil }).
		CheckStatus(StatusOK, StatusCreated, StatusNoContent, StatusBadRequest, StatusUnauthorized, StatusForbidden, StatusNotFound).
		Fetch(ctx)
	return status >= 200 && status < 500
}

// ---- 辅助函数：统一获取 BaseURL 与唯一值 ----
func GetBaseURLOrDefault(def string) string {
	v := os.Getenv("XCODING_BASE_URL")
	if strings.TrimSpace(v) != "" { return v }
	return def
}

func UniqueNano() int64 { return time.Now().UnixNano() }