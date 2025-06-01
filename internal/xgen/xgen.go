package xgen

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"x-gen/internal/captcha"
	"x-gen/internal/utils"
)

const (
	retryCount        = 3
	retryDelay        = 3 * time.Second
	maxWaitTime       = 60 * time.Second
	balanceCheckDelay = 5 * time.Second
)

type xGenerator struct {
	proxy      string
	captcha    *captcha.CaptchaServices
	httpClient *http.Client
}

func NewXGenerator(proxy string) *xGenerator {
	httpClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err == nil {
			transport := &http.Transport{
				Proxy:           http.ProxyURL(proxyUrl),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			httpClient.Transport = transport
		} else {
			utils.LogMessage(fmt.Sprintf("Invalid proxy URL: %v", err), "warning")
		}
	}
	return &xGenerator{
		proxy:      proxy,
		captcha:    captcha.NewCaptchaServicesWithProxy(proxy),
		httpClient: httpClient,
	}
}
