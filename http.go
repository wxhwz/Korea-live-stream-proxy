package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	retryAttempts  = 3               //最大重试次数
	retrySleepTime = time.Second * 2 //重试延迟时间
)

const (
	CapPacket      = false
	CapPacketProxy = "http://127.0.0.1:8888"
)

var Client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			if CapPacket {
				return url.Parse(CapPacketProxy)
			}
			return nil, nil // 不使用代理
		},
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		ForceAttemptHTTP2:   false,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
	},
}

// MRequest 发送 HTTP 请求并带重试机制
func MRequest(requestUrl, method string, requestBody interface{}, requestHeader map[string]string, isFormUrlEncoded bool) (int, string, error) {
	var lastErr error

	for attempt := 1; attempt <= retryAttempts; attempt++ {
		// 创建请求体
		var reqBody io.Reader
		if requestBody != nil {
			if isFormUrlEncoded {
				// 处理 x-www-form-urlencoded
				data := url.Values{}
				if bodyMap, ok := requestBody.(map[string]string); ok {
					for key, value := range bodyMap {
						data.Set(key, value)
					}
				}
				reqBody = strings.NewReader(data.Encode())
			} else {
				// 默认 JSON 处理
				jsonData, err := json.Marshal(requestBody)
				if err != nil {
					return 0, "", fmt.Errorf("attempt %d: json marshal failed: %w", attempt, err)
				}
				reqBody = bytes.NewBuffer([]byte(jsonData))
			}
		}

		// 创建请求
		req, err := http.NewRequest(strings.ToUpper(method), requestUrl, reqBody)
		if err != nil {
			return 0, "", fmt.Errorf("attempt %d: create request failed: %w", attempt, err)
		}

		// 设置请求头
		req.Header.Del("User-Agent") // 删除User-Agent头
		for key, value := range requestHeader {
			req.Header[key] = []string{value}
		}

		// 发送请求
		resp, err := Client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: request failed: %w", attempt, err)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return 0, "", lastErr
		}

		// 处理响应
		defer resp.Body.Close()

		// 处理 gzip 编码
		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("attempt %d: gzip decompression failed: %w", attempt, err)
				if attempt < retryAttempts {
					time.Sleep(retrySleepTime)
					continue
				}
				return 0, "", lastErr
			}
			reader = gzReader
			defer gzReader.Close()
		default:
			reader = resp.Body
		}

		// 读取响应体
		var respBody strings.Builder
		_, err = io.Copy(&respBody, reader)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: response reading failed: %w", attempt, err)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return 0, "", lastErr
		}

		// 如果状态码表示可重试的错误（如 429, 503），则继续重试
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lastErr = fmt.Errorf("attempt %d: server error with status code %d", attempt, resp.StatusCode)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return resp.StatusCode, respBody.String(), lastErr
		}

		return resp.StatusCode, respBody.String(), nil
	}

	return 0, "", lastErr
}

func MRequestCustom(requestUrl, method string, requestBody interface{}, requestHeader map[string]string, isFormUrlEncoded bool, retryAttempts int, retrySleepTime time.Duration) (int, string, error) {
	var lastErr error

	for attempt := 1; attempt <= retryAttempts; attempt++ {
		// 创建请求体
		var reqBody io.Reader
		if requestBody != nil {
			if isFormUrlEncoded {
				// 处理 x-www-form-urlencoded
				data := url.Values{}
				if bodyMap, ok := requestBody.(map[string]string); ok {
					for key, value := range bodyMap {
						data.Set(key, value)
					}
				}
				reqBody = strings.NewReader(data.Encode())
			} else {
				// 默认 JSON 处理
				jsonData, err := json.Marshal(requestBody)
				if err != nil {
					return 0, "", fmt.Errorf("attempt %d: json marshal failed: %w", attempt, err)
				}
				reqBody = bytes.NewBuffer([]byte(jsonData))
			}
		}

		// 创建请求
		req, err := http.NewRequest(strings.ToUpper(method), requestUrl, reqBody)
		if err != nil {
			return 0, "", fmt.Errorf("attempt %d: create request failed: %w", attempt, err)
		}

		// 设置请求头
		req.Header.Del("User-Agent") // 删除User-Agent头
		for key, value := range requestHeader {
			req.Header[key] = []string{value}
		}

		// 发送请求
		resp, err := Client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: request failed: %w", attempt, err)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return 0, "", lastErr
		}

		// 处理响应
		defer resp.Body.Close()

		// 处理 gzip 编码
		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("attempt %d: gzip decompression failed: %w", attempt, err)
				if attempt < retryAttempts {
					time.Sleep(retrySleepTime)
					continue
				}
				return 0, "", lastErr
			}
			reader = gzReader
			defer gzReader.Close()
		default:
			reader = resp.Body
		}

		// 读取响应体
		var respBody strings.Builder
		_, err = io.Copy(&respBody, reader)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: response reading failed: %w", attempt, err)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return 0, "", lastErr
		}

		// 如果状态码表示可重试的错误（如 429, 503），则继续重试
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lastErr = fmt.Errorf("attempt %d: server error with status code %d", attempt, resp.StatusCode)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return resp.StatusCode, respBody.String(), lastErr
		}

		return resp.StatusCode, respBody.String(), nil
	}

	return 0, "", lastErr
}

// MRequestTS 发送 TS 请求
func MRequestTS(requestUrl string, method string, requestHeader map[string]string) (*http.Response, error) {
	var lastErr error
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		// 创建请求
		req, err := http.NewRequest(strings.ToUpper(method), requestUrl, nil)
		if err != nil {
			return nil, fmt.Errorf("attempt %d: create request failed: %w", attempt, err)
		}

		// 设置请求头
		req.Header.Del("User-Agent") // 删除User-Agent头
		for key, value := range requestHeader {
			req.Header.Set(key, value)
		}

		// 发送请求
		resp, err := Client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: request failed: %w", attempt, err)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return nil, lastErr
		}

		// 如果状态码表示可重试的错误（如 429, 503），关闭响应体并重试
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("attempt %d: server error with status code %d", attempt, resp.StatusCode)
			if attempt < retryAttempts {
				time.Sleep(retrySleepTime)
				continue
			}
			return resp, lastErr
		}

		// 返回响应，调用者负责关闭 resp.Body
		return resp, nil
	}

	return nil, lastErr
}
