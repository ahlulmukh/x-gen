package xgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"x-gen/internal/utils"
)

const (
	xAuthToken = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
)

type XAccountInfo struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Username    string `json:"username"`
	UserId      string `json:"userId"`
	AuthToken   string `json:"authToken"`
	DateCreated string `json:"dateCreated"`
}

type XCookies struct {
	KDT               string
	GuestID           string
	GuestIDMarketing  string
	GuestIDAds        string
	CFBm              string
	PersonalizationID string
	GT                string
	TwitterSess       string
}

func (m *xGenerator) getDefaultHeaders(guestToken string) http.Header {
	headers := http.Header{}
	headers.Set("accept", "*/*")
	headers.Set("accept-language", "en-US,en;q=0.9")
	headers.Set("authorization", "Bearer "+xAuthToken)
	headers.Set("content-type", "application/json")
	headers.Set("priority", "u=1, i")
	headers.Set("sec-ch-ua", `"Chromium";v="136", "Google Chrome";v="136", "Not.A/Brand";v="99"`)
	headers.Set("sec-ch-ua-mobile", "?0")
	headers.Set("sec-ch-ua-platform", `"Windows"`)
	headers.Set("sec-fetch-dest", "empty")
	headers.Set("sec-fetch-mode", "cors")
	headers.Set("sec-fetch-site", "same-site")
	headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	headers.Set("referer", "https://x.com/")
	headers.Set("referrer-policy", "strict-origin-when-cross-origin")
	headers.Set("origin", "https://x.com")

	if guestToken != "" {
		headers.Set("x-guest-token", guestToken)
	}

	headers.Set("x-twitter-active-user", "yes")
	headers.Set("x-twitter-client-language", "en")

	return headers
}

func (m *xGenerator) generateTransactionId() string {
	const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	length := 86

	result := make([]byte, length)
	for i := range result {
		result[i] = characters[rand.Intn(len(characters))]
	}

	return string(result)
}

func (m *xGenerator) updateCookiesFromResponse(cookies *XCookies, resp *http.Response) {
	if resp == nil || resp.Header == nil {
		return
	}

	setCookieHeaders := resp.Header["Set-Cookie"]
	if len(setCookieHeaders) == 0 {
		return
	}

	utils.LogMessage(fmt.Sprintf("Received %d cookies from response", len(setCookieHeaders)), "info")

	for _, cookieStr := range setCookieHeaders {
		cookieParts := strings.Split(cookieStr, ";")
		if len(cookieParts) > 0 {
			nameValue := strings.Split(cookieParts[0], "=")
			if len(nameValue) >= 2 {
				name := strings.TrimSpace(nameValue[0])
				value := strings.TrimSpace(nameValue[1])

				switch name {
				case "kdt":
					cookies.KDT = value
				case "guest_id":
					cookies.GuestID = value
				case "guest_id_marketing":
					cookies.GuestIDMarketing = value
				case "guest_id_ads":
					cookies.GuestIDAds = value
				case "cf_bm":
					cookies.CFBm = value
				case "personalization_id":
					cookies.PersonalizationID = value
				case "gt":
					cookies.GT = value
				case "twitter_sess":
					cookies.TwitterSess = value
				}
			}
		}
	}
}

func (m *xGenerator) buildCookieString(cookies *XCookies) string {
	var cookieParts []string

	if cookies.KDT != "" {
		cookieParts = append(cookieParts, "kdt="+cookies.KDT)
	}
	if cookies.GuestID != "" {
		cookieParts = append(cookieParts, "guest_id="+cookies.GuestID)
	}
	if cookies.GuestIDMarketing != "" {
		cookieParts = append(cookieParts, "guest_id_marketing="+cookies.GuestIDMarketing)
	}
	if cookies.GuestIDAds != "" {
		cookieParts = append(cookieParts, "guest_id_ads="+cookies.GuestIDAds)
	}
	if cookies.CFBm != "" {
		cookieParts = append(cookieParts, "cf_bm="+cookies.CFBm)
	}
	if cookies.PersonalizationID != "" {
		cookieParts = append(cookieParts, "personalization_id="+cookies.PersonalizationID)
	}
	if cookies.GT != "" {
		cookieParts = append(cookieParts, "gt="+cookies.GT)
	}
	if cookies.TwitterSess != "" {
		cookieParts = append(cookieParts, "twitter_sess="+cookies.TwitterSess)
	}

	return strings.Join(cookieParts, "; ")
}

func (m *xGenerator) fetchGuestToken() (string, *XCookies, error) {
	utils.LogMessage("Fetching a valid guest token for X.com...", "process")

	cookies := &XCookies{}

	headers := m.getDefaultHeaders("")

	req, err := http.NewRequest("POST", "https://api.x.com/1.1/guest/activate.json", nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		utils.LogMessage("Failed to fetch guest token, trying fallback method...", "warning")
		return m.fetchGuestTokenFallback(cookies)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.LogMessage(fmt.Sprintf("Guest token request failed with status: %d", resp.StatusCode), "warning")
		return m.fetchGuestTokenFallback(cookies)
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		GuestToken string `json:"guest_token"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		utils.LogMessage("Failed to parse guest token response, trying fallback method...", "warning")
		return m.fetchGuestTokenFallback(cookies)
	}

	if response.GuestToken == "" {
		utils.LogMessage("Empty guest token received, trying fallback method...", "warning")
		return m.fetchGuestTokenFallback(cookies)
	}

	m.updateCookiesFromResponse(cookies, resp)

	cookies.GT = response.GuestToken
	cookies.GuestID = fmt.Sprintf("v1%%3A%d", time.Now().UnixNano()/int64(time.Millisecond))
	cookies.GuestIDMarketing = cookies.GuestID
	cookies.GuestIDAds = cookies.GuestID

	utils.LogMessage("Successfully obtained guest token", "success")
	return response.GuestToken, cookies, nil
}

func (m *xGenerator) fetchGuestTokenFallback(cookies *XCookies) (string, *XCookies, error) {
	utils.LogMessage("Trying fallback method to obtain guest token...", "process")

	req, err := http.NewRequest("GET", "https://x.com/", nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch X.com homepage: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to fetch X.com homepage: %d", resp.StatusCode)
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %v", err)
	}

	m.updateCookiesFromResponse(cookies, resp)

	re := regexp.MustCompile(`"gt=([0-9]+)"`)
	matches := re.FindStringSubmatch(string(body))

	if len(matches) < 2 || matches[1] == "" {
		return "", nil, fmt.Errorf("could not extract guest token from homepage")
	}

	extractedToken := matches[1]

	cookies.GT = extractedToken
	cookies.GuestID = fmt.Sprintf("v1%%3A%d", time.Now().UnixNano()/int64(time.Millisecond))
	cookies.GuestIDMarketing = cookies.GuestID
	cookies.GuestIDAds = cookies.GuestID

	utils.LogMessage("Successfully extracted guest token from homepage", "success")
	return extractedToken, cookies, nil
}

func (m *xGenerator) initiateSignupFlow(guestToken string, cookies *XCookies) (string, []byte, error) {
	utils.LogMessage("Initiating signup flow...", "process")

	headers := m.getDefaultHeaders(guestToken)
	headers.Set("x-client-transaction-id", m.generateTransactionId())

	cookieStr := m.buildCookieString(cookies)
	if cookieStr != "" {
		headers.Set("cookie", cookieStr)
	}

	requestBody := map[string]interface{}{
		"input_flow_data": map[string]interface{}{
			"requested_variant": `{"signup_type":"phone_email"}`,
			"flow_context": map[string]interface{}{
				"debug_overrides": map[string]interface{}{},
				"start_location": map[string]interface{}{
					"location": "splash_screen",
				},
			},
		},
		"subtask_versions": map[string]interface{}{
			"action_list":                          2,
			"alert_dialog":                         1,
			"app_download_cta":                     1,
			"check_logged_in_account":              1,
			"choice_selection":                     3,
			"contacts_live_sync_permission_prompt": 0,
			"cta":                                  7,
			"email_verification":                   2,
			"end_flow":                             1,
			"enter_date":                           1,
			"enter_email":                          2,
			"enter_password":                       5,
			"enter_phone":                          2,
			"enter_recaptcha":                      1,
			"enter_text":                           5,
			"enter_username":                       2,
			"generic_urt":                          3,
			"in_app_notification":                  1,
			"interest_picker":                      3,
			"js_instrumentation":                   1,
			"menu_dialog":                          1,
			"notifications_permission_prompt":      2,
			"open_account":                         2,
			"open_home_timeline":                   1,
			"open_link":                            1,
			"phone_verification":                   4,
			"privacy_options":                      1,
			"security_key":                         3,
			"select_avatar":                        4,
			"select_banner":                        2,
			"settings_list":                        7,
			"show_code":                            1,
			"sign_up":                              2,
			"sign_up_review":                       4,
			"tweet_selection_urt":                  1,
			"update_users":                         1,
			"upload_media":                         1,
			"user_recommendations_list":            4,
			"user_recommendations_urt":             1,
			"wait_spinner":                         3,
			"web_modal":                            1,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.x.com/1.1/onboarding/task.json?flow_name=signup", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	m.updateCookiesFromResponse(cookies, resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := readAll(resp.Body)
		return "", nil, fmt.Errorf("network response was not ok: %d - %s", resp.StatusCode, string(body))
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		FlowToken string `json:"flow_token"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.FlowToken == "" {
		return "", nil, fmt.Errorf("no flow token in response")
	}

	utils.LogMessage("Signup flow initiated successfully!", "success")
	return response.FlowToken, body, nil
}

func (m *xGenerator) checkEmailAvailability(email, guestToken string) (bool, error) {
	utils.LogMessage(fmt.Sprintf("Checking email availability for %s...", email), "process")

	headers := m.getDefaultHeaders(guestToken)
	headers.Set("x-client-transaction-id", m.generateTransactionId())

	encodedEmail := strings.Replace(url.QueryEscape(email), "+", "%20", -1)
	req, err := http.NewRequest("GET", "https://api.x.com/i/users/email_available.json?email="+encodedEmail, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := readAll(resp.Body)
		return false, fmt.Errorf("network response was not ok: %d - %s", resp.StatusCode, string(body))
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Valid bool `json:"valid"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return false, fmt.Errorf("failed to decode response: %v", err)
	}

	utils.LogMessage(
		fmt.Sprintf("Email availability check: %s", map[bool]string{true: "Available", false: "Not available"}[response.Valid]),
		"info")

	return response.Valid, nil
}

func (m *xGenerator) beginVerification(email, displayName, flowToken, guestToken string, cookies *XCookies) error {
	utils.LogMessage(fmt.Sprintf("Beginning verification for %s with name %s...", email, displayName), "process")

	headers := m.getDefaultHeaders(guestToken)
	headers.Set("x-client-transaction-id", m.generateTransactionId())

	cookieStr := m.buildCookieString(cookies)
	if cookieStr != "" {
		headers.Set("cookie", cookieStr)
	}

	requestBody := map[string]interface{}{
		"email":        email,
		"display_name": displayName,
		"flow_token":   flowToken,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.x.com/1.1/onboarding/begin_verification.json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	m.updateCookiesFromResponse(cookies, resp)

	if resp.StatusCode != http.StatusNoContent {
		body, _ := readAll(resp.Body)
		return fmt.Errorf("unexpected response: %d - %s", resp.StatusCode, string(body))
	}

	utils.LogMessage("Verification process initiated successfully!", "success")
	return nil
}

func (m *xGenerator) extractBlobFromTaskResponse(responseData []byte) string {

	var taskResponse map[string]interface{}
	if err := json.Unmarshal(responseData, &taskResponse); err != nil {
		utils.LogMessage(fmt.Sprintf("Failed to decode task response: %v", err), "error")
		return ""
	}

	subtasks, ok := taskResponse["subtasks"].([]interface{})
	if !ok {
		return ""
	}

	for _, subtask := range subtasks {
		subtaskMap, ok := subtask.(map[string]interface{})
		if !ok {
			continue
		}

		webModal, ok := subtaskMap["web_modal"].(map[string]interface{})
		if !ok {
			continue
		}

		urlStr, ok := webModal["url"].(string)
		if !ok {
			continue
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			utils.LogMessage(fmt.Sprintf("Failed to parse url %v", err), "error")
			continue
		}

		queryParams := parsedURL.Query()
		dataParam := queryParams.Get("data")
		if dataParam != "" {
			dataValue, err := url.QueryUnescape(dataParam)
			if err != nil {
				utils.LogMessage(fmt.Sprintf("Failed to decode parameter : %v", err), "error")
				continue
			}

			dataValue = strings.ReplaceAll(dataValue, " ", "+")
			preview := dataValue
			if len(dataValue) > 50 {
				preview = dataValue[:50] + "..."
			}
			utils.LogMessage(fmt.Sprintf("Paramter Data Succesfully Extracted: %s", preview), "success")
			return dataValue
		}
	}

	return ""
}

func (m *xGenerator) submitUserInfo(name, email, verificationCode, flowToken, guestToken, captchaToken string, cookies *XCookies) (string, map[string]interface{}, error) {
	utils.LogMessage("Submitting user information...", "process")

	headers := m.getDefaultHeaders(guestToken)
	headers.Set("x-client-transaction-id", m.generateTransactionId())

	cookieStr := m.buildCookieString(cookies)
	if cookieStr != "" {
		headers.Set("cookie", cookieStr)
	}

	currentYear := time.Now().Year()
	year := currentYear - 22 - rand.Intn(18)
	month := 1 + rand.Intn(12)
	day := 1 + rand.Intn(28)

	requestBody := map[string]interface{}{
		"flow_token": flowToken,
		"subtask_inputs": []map[string]interface{}{
			{
				"subtask_id": "Signup",
				"sign_up": map[string]interface{}{
					"link":  "email_next_link",
					"name":  name,
					"email": email,
					"birthday": map[string]interface{}{
						"day":   day,
						"month": month,
						"year":  year,
					},
					"personalization_settings": map[string]interface{}{
						"allow_cookie_use":             false,
						"allow_device_personalization": false,
						"allow_partnerships":           false,
						"allow_ads_personalization":    false,
					},
				},
			},
			{
				"subtask_id": "ArkoseEmail",
				"web_modal": map[string]interface{}{
					"completion_deeplink": fmt.Sprintf("twitter://onboarding/web_modal/next_link?access_token=%s", captchaToken),
					"link":                "signup_with_email_next_link",
				},
			},
			{
				"subtask_id": "EmailVerification",
				"email_verification": map[string]interface{}{
					"code":  verificationCode,
					"email": email,
					"link":  "next_link",
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.x.com/1.1/onboarding/task.json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	m.updateCookiesFromResponse(cookies, resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := readAll(resp.Body)
		return "", nil, fmt.Errorf("network response was not ok: %d - %s", resp.StatusCode, string(body))
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %v", err)
	}

	flowToken, ok := response["flow_token"].(string)
	if !ok || flowToken == "" {
		return "", nil, fmt.Errorf("no flow token in response")
	}

	utils.LogMessage("User information submitted successfully!", "success")
	return flowToken, response, nil
}

func (m *xGenerator) checkPasswordStrength(password, username, guestToken string) (bool, error) {
	utils.LogMessage("Checking password strength...", "process")

	headers := m.getDefaultHeaders(guestToken)
	headers.Set("content-type", "application/x-www-form-urlencoded")
	headers.Set("x-client-transaction-id", m.generateTransactionId())

	formValues := fmt.Sprintf("password=%s&username=%s", url.QueryEscape(password), url.QueryEscape(username))

	req, err := http.NewRequest("POST", "https://api.x.com/1.1/account/password_strength.json", strings.NewReader(formValues))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := readAll(resp.Body)
		return false, fmt.Errorf("network response was not ok: %d - %s", resp.StatusCode, string(body))
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Pass bool `json:"pass"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return false, fmt.Errorf("failed to decode response: %v", err)
	}

	utils.LogMessage(
		fmt.Sprintf("Password strength check: %s", map[bool]string{true: "Passed", false: "Failed"}[response.Pass]),
		"info")

	return response.Pass, nil
}

func (m *xGenerator) completeSignup(password, flowToken, guestToken string, cookies *XCookies) (string, string, string, error) {
	utils.LogMessage("Completing signup with password...", "process")

	headers := m.getDefaultHeaders(guestToken)
	headers.Set("x-client-transaction-id", m.generateTransactionId())

	cookieStr := m.buildCookieString(cookies)
	if cookieStr != "" {
		headers.Set("cookie", cookieStr)
	}

	requestBody := map[string]interface{}{
		"flow_token": flowToken,
		"subtask_inputs": []map[string]interface{}{
			{
				"subtask_id": "EnterPassword",
				"enter_password": map[string]interface{}{
					"password": password,
					"link":     "next_link",
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.x.com/1.1/onboarding/task.json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %v", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	m.updateCookiesFromResponse(cookies, resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := readAll(resp.Body)
		return "", "", "", fmt.Errorf("network response was not ok: %d - %s", resp.StatusCode, string(body))
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read response body: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	subtasks, ok := response["subtasks"].([]interface{})
	if !ok || len(subtasks) == 0 {
		return "", "", "", fmt.Errorf("no subtasks in response")
	}

	var userId, screenName, authToken string

	if firstSubtask, ok := subtasks[0].(map[string]interface{}); ok {
		if openAccount, ok := firstSubtask["open_account"].(map[string]interface{}); ok {

			authToken, _ = openAccount["auth_token"].(string)

			if user, ok := openAccount["user"].(map[string]interface{}); ok {

				userId, ok = user["id_str"].(string)
				if !ok || userId == "" {

					if idFloat, ok := user["id"].(float64); ok {
						userId = fmt.Sprintf("%.0f", idFloat)
					}
				}

				screenName, _ = user["screen_name"].(string)
			}
		}
	}

	if userId == "" || screenName == "" || authToken == "" {
		utils.LogMessage("Primary extraction method failed, trying alternative methods...", "warning")

		for _, subtask := range subtasks {
			subtaskMap, ok := subtask.(map[string]interface{})
			if !ok {
				continue
			}

			if user, ok := subtaskMap["user"].(map[string]interface{}); ok {
				if userId == "" {
					if idStr, ok := user["id_str"].(string); ok {
						userId = idStr
					} else if idFloat, ok := user["id"].(float64); ok {
						userId = fmt.Sprintf("%.0f", idFloat)
					}
				}
				if screenName == "" {
					screenName, _ = user["screen_name"].(string)
				}
			}

			if authToken == "" {
				authToken, _ = subtaskMap["auth_token"].(string)
			}

			for _, value := range subtaskMap {
				if valueMap, ok := value.(map[string]interface{}); ok {
					if authToken == "" {
						if token, ok := valueMap["auth_token"].(string); ok {
							authToken = token
						}
					}

					if user, ok := valueMap["user"].(map[string]interface{}); ok {
						if userId == "" {
							if idStr, ok := user["id_str"].(string); ok {
								userId = idStr
							} else if idFloat, ok := user["id"].(float64); ok {
								userId = fmt.Sprintf("%.0f", idFloat)
							}
						}
						if screenName == "" {
							screenName, _ = user["screen_name"].(string)
						}
					}
				}
			}
		}
	}

	if userId == "" || screenName == "" || authToken == "" {
		utils.LogMessage("Failed to extract all required user information", "error")
		return "", "", "", fmt.Errorf("missing user information in response")
	}

	utils.LogMessage(
		fmt.Sprintf("Account created successfully! Username: @%s, User ID: %s", screenName, userId),
		"success")

	return userId, screenName, authToken, nil
}

func (m *xGenerator) GenerateXAccount() (*XAccountInfo, error) {
	utils.LogMessage("Starting X.com account generation process...", "process")

	var guestToken string
	var cookies *XCookies
	var err error

	rand.Seed(time.Now().UnixNano())

	var mailHandler *utils.MailTmHandler
	utils.LogMessage("Creating temporary email using mail.tm...", "process")

	mailHandler = utils.NewMailTmHandler(m.proxy)
	password := utils.GeneratePassword()
	name := utils.GenerateName()

	err = mailHandler.CreateMailAccount(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary email account: %v", err)
	}
	email := mailHandler.GetMailAddress()
	utils.LogMessage(fmt.Sprintf("Using temporary email: %s", email), "success")
	utils.LogMessage(fmt.Sprintf("Generated name: %s", name), "success")

	retries := 0
	for retries < retryCount {
		guestToken, cookies, err = m.fetchGuestToken()
		if err == nil {
			break
		}

		retries++
		utils.LogMessage(
			fmt.Sprintf("Attempt %d/%d failed to get guest token: %v", retries, retryCount, err),
			"warning")

		if retries >= retryCount {
			return nil, fmt.Errorf("failed to obtain a guest token after multiple attempts")
		}

		time.Sleep(retryDelay * time.Duration(retries))
	}

	retries = 0
	var flowToken string
	var responseData []byte
	var blobData string

	for retries < retryCount {
		flowToken, responseData, err = m.initiateSignupFlow(guestToken, cookies)
		if err == nil {
			blobData = m.extractBlobFromTaskResponse(responseData)
			if blobData != "" {
				utils.LogMessage("Successfully extracted blob data for FunCaptcha!", "info")
				// captchaToken, err := m.captcha.SolveCaptcha(m.currentNum, m.total, blobData)
				// if err != nil {
				// 	return nil, fmt.Errorf("failed to solve captcha: %v", err)
				// }
				// print("Captcha token: ")
				// fmt.Println(captchaToken)
			}
			break
		}

		retries++
		utils.LogMessage(
			fmt.Sprintf("Signup initiation attempt %d/%d failed: %v", retries, retryCount, err),
			"warning")

		if strings.Contains(err.Error(), "403") {
			utils.LogMessage("Trying to obtain a new guest token...", "process")
			guestToken, cookies, err = m.fetchGuestToken()
			if err != nil {
				utils.LogMessage(fmt.Sprintf("Failed to get a new guest token: %v", err), "error")
			} else {
				utils.LogMessage("Successfully obtained a new guest token", "success")
			}
		}

		if retries >= retryCount {
			return nil, fmt.Errorf("failed to initiate signup flow after multiple attempts")
		}

		time.Sleep(retryDelay * time.Duration(retries))
	}

	isEmailAvailable, err := m.checkEmailAvailability(email, guestToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check email availability: %v", err)
	}

	if !isEmailAvailable {
		return nil, fmt.Errorf("email is already registered or invalid")
	}

	err = m.beginVerification(email, name, flowToken, guestToken, cookies)
	if err != nil {
		return nil, fmt.Errorf("failed to begin verification: %v", err)
	}

	utils.LogMessage("Getting verification code...", "process")
	var verificationCode string

	if mailHandler != nil {
		code, err := mailHandler.WaitForVerificationEmail(15)
		if err != nil {
			utils.LogMessage(fmt.Sprintf("Failed to get verification code automatically: %v", err), "error")
			utils.LogMessage("Falling back to manual verification...", "info")
			fmt.Print("Enter verification code: ")
			fmt.Scan(&verificationCode)
		} else {
			verificationCode = code
			utils.LogMessage(fmt.Sprintf("Verification code obtained automatically: %s", verificationCode), "success")
		}
	}

	captchaToken, err := m.captcha.SolveCaptcha(blobData)
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Failed to solve captcha: %v", err), "error")
		return nil, fmt.Errorf("failed to solve captcha: %v", err)
	}

	newFlowToken, userInfoData, err := m.submitUserInfo(name, email, verificationCode, flowToken, guestToken, captchaToken, cookies)
	if err != nil {
		return nil, fmt.Errorf("failed to submit user information: %v", err)
	}

	var username string
	if subtasks, ok := userInfoData["subtasks"].([]interface{}); ok && len(subtasks) > 0 {
		if enterPassword, ok := subtasks[0].(map[string]interface{})["enter_password"].(map[string]interface{}); ok {
			username, _ = enterPassword["username"].(string)
		}
	}

	if username == "" {
		utils.LogMessage("Could not extract assigned username, using email prefix instead", "warning")
		username = strings.Split(email, "@")[0]
	} else {
		utils.LogMessage(fmt.Sprintf("Assigned username: @%s", username), "info")
	}

	isPasswordStrong, err := m.checkPasswordStrength(password, username, guestToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check password strength: %v", err)
	}

	if !isPasswordStrong {
		return nil, fmt.Errorf("password is not strong enough")
	}

	userId, screenName, authToken, err := m.completeSignup(password, newFlowToken, guestToken, cookies)
	if err != nil {
		return nil, fmt.Errorf("failed to complete signup: %v", err)
	}

	accountInfo := &XAccountInfo{
		Name:        name,
		Email:       email,
		Password:    password,
		Username:    screenName,
		UserId:      userId,
		AuthToken:   authToken,
		DateCreated: time.Now().Format(time.RFC3339),
	}

	err = utils.SaveTokenFile(authToken)
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Failed to save account to authtoken.txt: %v", err), "warning")
	} else {
		utils.LogMessage("Account saved to authtoken.txt", "success")
	}

	err = utils.SaveAccountToFile(email, password, screenName)
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Failed to save account to accounts.txt: %v", err), "warning")
	} else {
		utils.LogMessage("Account saved to accounts.txt", "success")
	}

	utils.LogMessage("X.com account creation completed successfully!", "success")
	return accountInfo, nil
}

func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
