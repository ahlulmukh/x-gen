package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type MailTmHandler struct {
	client       *http.Client
	apiBaseURL   string
	mailAddress  string
	mailToken    string
	mailPassword string
	proxy        string
}

type MailTmDomain struct {
	ID        string `json:"id"`
	Domain    string `json:"domain"`
	IsActive  bool   `json:"isActive"`
	IsPrivate bool   `json:"isPrivate"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type MailTmDomainsResponse struct {
	HydraMember []MailTmDomain `json:"hydra:member"`
}

type MailTmAccount struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

type MailTmMessage struct {
	ID      string     `json:"id"`
	From    MailTmFrom `json:"from"`
	To      []MailTmTo `json:"to"`
	Subject string     `json:"subject"`
	Intro   string     `json:"intro"`
	Text    string     `json:"text"`
	HTML    string     `json:"html"`
}

type MailTmFrom struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type MailTmTo struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type MailTmMessagesResponse struct {
	HydraMember []MailTmMessage `json:"hydra:member"`
}

type MailTmTokenResponse struct {
	Token string `json:"token"`
}

func NewMailTmHandler(proxyStr string) *MailTmHandler {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if proxyStr != "" {
		proxyURL, err := url.Parse(proxyStr)
		if err == nil {
			transport := &http.Transport{
				Proxy:               http.ProxyURL(proxyURL),
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				DisableKeepAlives:   true,
				MaxIdleConnsPerHost: -1,
			}
			client.Transport = transport
		} else {
			LogMessage(fmt.Sprintf("Invalid mail proxy URL: %v", err), "warning")
		}
	}

	return &MailTmHandler{
		client:     client,
		apiBaseURL: "https://api.mail.tm",
		proxy:      proxyStr,
	}
}

func (m *MailTmHandler) GetAvailableDomains() (string, error) {
	LogMessage("Fetching available Mail.tm domains...", "info")

	resp, err := m.client.Get(fmt.Sprintf("%s/domains", m.apiBaseURL))
	if err != nil {
		return "", fmt.Errorf("failed to fetch domains: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get domains: HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var domains MailTmDomainsResponse
	if err := json.Unmarshal(body, &domains); err != nil {
		return "", fmt.Errorf("failed to parse domains response: %v", err)
	}

	if len(domains.HydraMember) == 0 {
		return "", fmt.Errorf("no domains available")
	}

	for _, domain := range domains.HydraMember {
		if domain.IsActive {
			return domain.Domain, nil
		}
	}

	return "", fmt.Errorf("no active domains found")
}

func (m *MailTmHandler) GenerateRandomUsername(length int) string {
	if length < 5 {
		length = 10
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	const letters = "abcdefghijklmnopqrstuvwxyz"
	const digits = "0123456789"

	username := make([]byte, length)

	username[0] = letters[r.Intn(len(letters))]

	for i := 1; i < length; i++ {
		if r.Intn(3) == 0 {
			username[i] = digits[r.Intn(len(digits))]
		} else {
			username[i] = letters[r.Intn(len(letters))]
		}
	}

	for i := range username {
		j := r.Intn(i + 1)
		username[i], username[j] = username[j], username[i]
	}

	return string(username)
}

func (m *MailTmHandler) CreateMailAccount(password string) error {
	domain, err := m.GetAvailableDomains()
	if err != nil {
		return err
	}

	username := m.GenerateRandomUsername(12)
	m.mailAddress = fmt.Sprintf("%s@%s", username, domain)
	m.mailPassword = password

	LogMessage(fmt.Sprintf("Creating Mail.tm account: %s", m.mailAddress), "info")

	account := MailTmAccount{
		Address:  m.mailAddress,
		Password: m.mailPassword,
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account data: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts", m.apiBaseURL), bytes.NewBuffer(accountJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create account: HTTP status %d - %s", resp.StatusCode, string(body))
	}

	LogMessage("Mail.tm account created successfully", "success")
	return m.LoginMailAccount()
}

func (m *MailTmHandler) LoginMailAccount() error {
	LogMessage(fmt.Sprintf("Logging into Mail.tm: %s", m.mailAddress), "info")

	account := MailTmAccount{
		Address:  m.mailAddress,
		Password: m.mailPassword,
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal login data: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/token", m.apiBaseURL), bytes.NewBuffer(accountJSON))
	if err != nil {
		return fmt.Errorf("failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: HTTP status %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %v", err)
	}

	var tokenResp MailTmTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	m.mailToken = tokenResp.Token
	LogMessage("Successfully logged into Mail.tm account", "success")
	return nil
}

func (m *MailTmHandler) GetAllMessages() ([]MailTmMessage, error) {
	if m.mailToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	LogMessage("Checking for incoming emails...", "info")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/messages", m.apiBaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.mailToken))

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		LogMessage(fmt.Sprintf("Rate limited (429) by mail.tm. Retry-After: %s. Proxy: %v", retryAfter, m.proxy != ""), "warning")
		time.Sleep(5 * time.Second)
		return nil, fmt.Errorf("rate limited (429) by mail.tm")
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get messages: HTTP status %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var messagesResp MailTmMessagesResponse
	if err := json.Unmarshal(body, &messagesResp); err != nil {
		return nil, fmt.Errorf("failed to parse messages response: %v", err)
	}
	return messagesResp.HydraMember, nil
}

func (m *MailTmHandler) WaitForVerificationEmail(maxAttempts int) (string, error) {
	if maxAttempts <= 0 {
		maxAttempts = 10
	}

	LogMessage("Waiting for verification email...", "info")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		messages, err := m.GetAllMessages()
		if err != nil {
			LogMessage(fmt.Sprintf("Error checking emails: %v", err), "warning")
		} else if len(messages) > 0 {
			LogMessage(fmt.Sprintf("Found %d messages", len(messages)), "info")

			for _, msg := range messages {

				if strings.Contains(msg.Subject, "verification") ||
					strings.Contains(msg.Subject, "Verification") ||
					strings.Contains(msg.Subject, "code") ||
					strings.Contains(msg.Subject, "Code") ||
					strings.Contains(msg.Subject, "X.com") ||
					strings.Contains(msg.Subject, "X verification") ||
					strings.Contains(msg.Subject, "Twitter") {

					code := ExtractVerificationCodeFromSubject(msg.Subject)
					if code != "" {
						return code, nil
					}

					LogMessage(fmt.Sprintf("Found potential verification email. Text content length: %d, HTML content length: %d",
						len(msg.Text), len(msg.HTML)), "debug")

					if msg.Text == "" && msg.HTML == "" {
						LogMessage("Warning: Message content is empty, but already checked subject line", "info")
						continue
					}

					codeFromText := extractVerificationCode(msg.Text)
					if codeFromText != "" {
						return codeFromText, nil
					}

					codeFromHTML := extractVerificationCode(msg.HTML)
					if codeFromHTML != "" {
						return codeFromHTML, nil
					}

					re := regexp.MustCompile(`\b\d{6}\b`)
					matchStr := re.FindString(msg.Text)
					if matchStr != "" {
						LogMessage(fmt.Sprintf("Found verification code with brute force approach: %s", matchStr), "success")
						return matchStr, nil
					}

					LogMessage("Found verification email but couldn't automatically extract code", "warning")
					return "", fmt.Errorf("verification email found but code extraction failed, message text: %s", msg.Text)
				}
			}
		}

		time.Sleep(5 * time.Second)
	}

	return "", fmt.Errorf("no verification email received after %d attempts", maxAttempts)
}

func extractVerificationCode(content string) string {

	if content == "" {
		LogMessage("Content is empty, unable to extract verification code", "warning")
		return ""
	}

	patterns := []string{

		`verification code to get started on X:[\s\n]*(\d{6})`,
		`verification code to get started on X is:[\s\n]*(\d{6})`,
		`Please enter this verification code to get started on X:[\s\n]*(\d{6})`,
		`Please enter this verification code[\s\S]*?:[\s\n]*(\d{6})`,
		`Your X verification code is (\d{6})`,
		`Your Twitter verification code is (\d{6})`,
		`Your verification code is (\d{6})`,

		`verification code is (\d{4,8})`,
		`verification code: (\d{4,8})`,
		`verification code[\s\S]*?:[\s\n]*(\d{4,8})`,
		`code is (\d{4,8})`,
		`code: (\d{4,8})`,

		`[\s\n](\d{6})[\s\n]`,
		`(\d{6})`,
	}

	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) >= 2 {
			LogMessage(fmt.Sprintf("Found code with pattern #%d: %s", i+1, matches[1]), "success")
			return matches[1]
		}
	}

	re := regexp.MustCompile(`\D(\d{6})\D`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 2 {
		LogMessage(fmt.Sprintf("Found 6-digit code with general pattern: %s", matches[1]), "success")
		return matches[1]
	}

	return ""
}

func (m *MailTmHandler) GetMailAddress() string {
	return m.mailAddress
}

func (m *MailTmHandler) GetMailPassword() string {
	return m.mailPassword
}

func (m *MailTmHandler) GetSpecificMessage(messageID string) (*MailTmMessage, error) {
	if m.mailToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	LogMessage(fmt.Sprintf("Fetching message with ID: %s", messageID), "info")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/messages/%s", m.apiBaseURL, messageID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.mailToken))

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get message: HTTP status %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var message MailTmMessage
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("failed to parse message response: %v", err)
	}

	return &message, nil
}

func ExtractVerificationCodeFromSubject(subject string) string {

	patterns := []string{
		`^(\d{6}) is your X verification code$`,
		`^(\d{6}) is your X verification code\.?$`,
		`^Your X verification code is (\d{6})$`,
		`^Your X verification code is (\d{6})\.?$`,
		`^X verification code: (\d{6})$`,
		`^Twitter verification code: (\d{6})$`,
		`^Twitter code: (\d{6})$`,
		`^Confirmation code: (\d{6})$`,
		`^(\d{6}) is your Twitter verification code$`,
		`^(\d{6})$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(subject)
		if len(matches) >= 2 {

			return matches[1]
		}
	}

	re := regexp.MustCompile(`\b(\d{6})\b`)
	matches := re.FindStringSubmatch(subject)
	if len(matches) >= 2 {
		LogMessage(fmt.Sprintf("Found 6-digit code in subject with general regex: %s", matches[1]), "success")
		return matches[1]
	}

	return ""

}
