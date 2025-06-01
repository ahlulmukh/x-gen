package captcha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"x-gen/internal/utils"
)

type CaptchaServices struct {
	sitekey           string
	pageUrl           string
	antiCaptchaApiUrl string
	twocaptchaApiUrl  string
	proxy             string
}

type Config struct {
	CaptchaServices struct {
		CaptchaUsing      string   `json:"captchaUsing"`
		UrlPrivate        string   `json:"urlPrivate"`
		AntiCaptchaApikey []string `json:"antiCaptchaApikey"`
		Captcha2Apikey    []string `json:"captcha2Apikey"`
	} `json:"captchaServices"`
}

func NewCaptchaServices() *CaptchaServices {
	return &CaptchaServices{
		sitekey:           "2CB16598-CB82-4CF7-B332-5990DB66F3AB",
		pageUrl:           "https://x.com",
		antiCaptchaApiUrl: "https://api.anti-captcha.com",
		twocaptchaApiUrl:  "https://api.2captcha.com",
		proxy:             "",
	}
}

func NewCaptchaServicesWithProxy(proxy string) *CaptchaServices {
	return &CaptchaServices{
		sitekey:           "2CB16598-CB82-4CF7-B332-5990DB66F3AB",
		pageUrl:           "https://x.com",
		antiCaptchaApiUrl: "https://api.anti-captcha.com",
		twocaptchaApiUrl:  "https://api.2captcha.com",
		proxy:             proxy,
	}
}

func LoadConfig() *Config {
	file, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

func (cs *CaptchaServices) parseProxy() (proxyType, proxyAddress, proxyLogin, proxyPassword string, proxyPort int) {
	if cs.proxy == "" {
		return "", "", "", "", 0
	}

	proxyURL, err := url.Parse(cs.proxy)
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Failed to parse proxy URL: %v", err), "error")
		return "", "", "", "", 0
	}

	switch strings.ToLower(proxyURL.Scheme) {
	case "http", "https":
		proxyType = "http"
	case "socks4":
		proxyType = "socks4"
	case "socks5":
		proxyType = "socks5"
	default:
		proxyType = "http"
	}

	proxyAddress = proxyURL.Hostname()
	if proxyURL.Port() != "" {
		fmt.Sscanf(proxyURL.Port(), "%d", &proxyPort)
	}

	if proxyURL.User != nil {
		proxyLogin = proxyURL.User.Username()
		proxyPassword, _ = proxyURL.User.Password()
	}

	return proxyType, proxyAddress, proxyLogin, proxyPassword, proxyPort
}

func (cs *CaptchaServices) SolveCaptcha(blob string) (string, error) {
	config := LoadConfig()
	provider := config.CaptchaServices.CaptchaUsing

	maxRetries := 5
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		utils.LogMessage(fmt.Sprintf("Captcha solve attempt %d/%d...", attempt, maxRetries), "process")

		var result string
		var err error

		switch provider {
		case "2captcha":
			result, err = cs.solveCaptcha2(blob)
		case "antiCaptcha":
			result, err = cs.antiCaptcha(blob)
		case "private":
			result, err = cs.solvedPrivate(blob)
		default:
			utils.LogMessage("Invalid captcha provider.", "error")
			return "", fmt.Errorf("invalid captcha provider")
		}

		if err == nil && result != "" {
			utils.LogMessage(fmt.Sprintf("Captcha solved successfully on attempt %d!", attempt), "success")
			return result, nil
		}

		utils.LogMessage(fmt.Sprintf("Captcha solve attempt %d failed: %v", attempt, err), "warning")

		if attempt < maxRetries {
			utils.LogMessage(fmt.Sprintf("Waiting %v before retry...", retryDelay), "info")
			time.Sleep(retryDelay)
			retryDelay = retryDelay * 2
		}
	}

	return "", fmt.Errorf("failed to solve captcha after %d attempts", maxRetries)
}

func (cs *CaptchaServices) solvedPrivate(blob string) (string, error) {
	utils.LogMessage("Trying to solve captcha using local service...", "process")

	createTaskData := map[string]interface{}{
		"preset":         "twitter_register",
		"chrome_version": "136",
		"proxy":          "",
		"blob":           blob,
	}

	jsonData, err := json.Marshal(createTaskData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal create task data: %v", err)
	}
	resp, err := http.Post("http://127.0.0.1:8003/funcaptcha/createTask", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create task: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read create task response: %v", err)
	}

	var createTaskResult struct {
		Success bool   `json:"success"`
		TaskID  string `json:"task_id"`
	}

	if err := json.Unmarshal(body, &createTaskResult); err != nil {
		return "", fmt.Errorf("failed to parse create task response: %v", err)
	}

	if !createTaskResult.Success {
		return "", fmt.Errorf("task creation failed")
	}

	taskID := createTaskResult.TaskID
	utils.LogMessage(fmt.Sprintf("Captcha task created with ID: %s", taskID), "process")

	getTaskData := map[string]interface{}{
		"task_id": taskID,
	}

	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)

		jsonData, err = json.Marshal(getTaskData)
		if err != nil {
			continue
		}
		resp, err = http.Post("http://127.0.0.1:8003/funcaptcha/getTask", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			continue
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			continue
		}

		var taskResult map[string]interface{}
		if err := json.Unmarshal(body, &taskResult); err != nil {
			continue
		}

		status, ok := taskResult["status"].(string)
		if !ok {
			continue
		}

		if status == "processing" {
			utils.LogMessage("Captcha still processing...", "info")
			continue
		}

		if status == "completed" {
			captcha, ok := taskResult["captcha"].(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("invalid response format - missing captcha field")
			}

			success, ok := captcha["success"].(bool)
			if !ok || !success {
				errMsg := "Unknown error"
				if errStr, ok := captcha["err"].(string); ok {
					errMsg = errStr
				}
				return "", fmt.Errorf("captcha solving failed: %s", errMsg)
			}

			token, ok := captcha["token"].(string)
			if !ok || token == "" {
				return "", fmt.Errorf("no token in response")
			}

			utils.LogMessage("Captcha solved successfully!", "success")
			return token, nil
		}
	}

	return "", fmt.Errorf("timed out waiting for captcha solution")
}

func (cs *CaptchaServices) antiCaptcha(blob string) (string, error) {
	utils.LogMessage("Trying solving Fun Captcha with anticaptcha ...", "process")
	config := LoadConfig()
	apiKey := config.CaptchaServices.AntiCaptchaApikey[0]

	formattedBlob := fmt.Sprintf("{\"blob\":\"%s\"}", blob)
	proxyType, proxyAddress, proxyLogin, proxyPassword, proxyPort := cs.parseProxy()

	taskData := map[string]interface{}{
		"clientKey": apiKey,
		"task": map[string]interface{}{
			"type":                     "FunCaptchaTask",
			"websiteURL":               cs.pageUrl,
			"websitePublicKey":         cs.sitekey,
			"funcaptchaApiJSSubdomain": "client-api.arkoselabs.com",
			"data":                     formattedBlob,
			"userAgent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		},
		"softId": 0,
	}

	if proxyAddress != "" {
		task := taskData["task"].(map[string]interface{})
		task["proxyType"] = proxyType
		task["proxyAddress"] = proxyAddress
		task["proxyPort"] = proxyPort
		task["proxyLogin"] = proxyLogin
		task["proxyPassword"] = proxyPassword

		utils.LogMessage(fmt.Sprintf("Using proxy: %s:%d", proxyAddress, proxyPort), "info")
	} else {
		task := taskData["task"].(map[string]interface{})
		task["type"] = "FunCaptchaTaskProxyless"
		utils.LogMessage("No proxy configured, using proxyless mode", "info")
	}
	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(cs.antiCaptchaApiUrl+"/createTask", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var createTaskResp struct {
		ErrorId          int    `json:"errorId"`
		TaskId           int    `json:"taskId"`
		Status           string `json:"status"`
		ErrorCode        string `json:"errorCode"`
		ErrorDescription string `json:"errorDescription"`
	}

	if err := json.Unmarshal(body, &createTaskResp); err != nil {
		utils.LogMessage(fmt.Sprintf("Failed to parse response: %s", string(body)), "error")
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if createTaskResp.ErrorId != 0 || createTaskResp.TaskId == 0 {
		errorMsg := fmt.Sprintf("AntiCaptcha API Error - ID: %d, Code: %s, Description: %s",
			createTaskResp.ErrorId, createTaskResp.ErrorCode, createTaskResp.ErrorDescription)
		utils.LogMessage(errorMsg, "error")
		utils.LogMessage(fmt.Sprintf("Full response: %s", string(body)), "error")
		return "", fmt.Errorf("failed to create task: %s", errorMsg)
	}

	utils.LogMessage(fmt.Sprintf("Task created with ID: %d", createTaskResp.TaskId), "process")

	getTaskData := map[string]interface{}{
		"clientKey": apiKey,
		"taskId":    createTaskResp.TaskId,
	}

	var result string
	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)

		jsonData, err = json.Marshal(getTaskData)
		if err != nil {
			continue
		}

		resp, err = http.Post(cs.antiCaptchaApiUrl+"/getTaskResult", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			continue
		}

		var taskResult struct {
			ErrorId  int    `json:"errorId"`
			Status   string `json:"status"`
			Solution struct {
				Token string `json:"token"`
			} `json:"solution"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&taskResult); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		if taskResult.Status == "ready" {
			result = taskResult.Solution.Token
			utils.LogMessage("Captcha solved successfully!", "success")
			break
		}
	}

	if result == "" {
		return "", fmt.Errorf("failed to get captcha solution")
	}

	return result, nil
}

func (cs *CaptchaServices) solveCaptcha2(blob string) (string, error) {
	utils.LogMessage("Trying solving Fun captcha with 2 captcha ...", "process")
	config := LoadConfig()
	apiKey := config.CaptchaServices.Captcha2Apikey[0]

	formattedBlob := fmt.Sprintf("{\"blob\":\"%s\"}", blob)

	proxyType, proxyAddress, proxyLogin, proxyPassword, proxyPort := cs.parseProxy()

	taskData := map[string]interface{}{
		"clientKey": apiKey,
		"task": map[string]interface{}{
			"type":                     "FunCaptchaTask",
			"websiteURL":               cs.pageUrl,
			"websitePublicKey":         cs.sitekey,
			"funcaptchaApiJSSubdomain": "client-api.arkoselabs.com",
			"data":                     formattedBlob,
			"userAgent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		},
		"softId": 0,
	}
	if proxyAddress != "" {
		task := taskData["task"].(map[string]interface{})
		task["proxyType"] = proxyType
		task["proxyAddress"] = proxyAddress
		task["proxyPort"] = proxyPort
		task["proxyLogin"] = proxyLogin
		task["proxyPassword"] = proxyPassword

		utils.LogMessage(fmt.Sprintf("Using proxy: %s:%d", proxyAddress, proxyPort), "info")
	} else {
		task := taskData["task"].(map[string]interface{})
		task["type"] = "FunCaptchaTaskProxyless"
		utils.LogMessage("No proxy configured, using proxyless mode", "info")
	}
	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return "", err
	}

	utils.LogMessage(fmt.Sprintf("Sending request to 2Captcha: %s", string(jsonData)), "info")

	resp, err := http.Post(cs.twocaptchaApiUrl+"/createTask", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var createTaskResp struct {
		ErrorId          int    `json:"errorId"`
		TaskId           int    `json:"taskId"`
		Status           string `json:"status"`
		ErrorCode        string `json:"errorCode"`
		ErrorDescription string `json:"errorDescription"`
	}
	if err := json.Unmarshal(body, &createTaskResp); err != nil {
		utils.LogMessage(fmt.Sprintf("Failed to parse response: %s", string(body)), "error")
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if createTaskResp.ErrorId != 0 || createTaskResp.TaskId == 0 {
		errorMsg := fmt.Sprintf("2Captcha API Error - ID: %d, Code: %s, Description: %s",
			createTaskResp.ErrorId, createTaskResp.ErrorCode, createTaskResp.ErrorDescription)
		utils.LogMessage(errorMsg, "error")
		utils.LogMessage(fmt.Sprintf("Full response: %s", string(body)), "error")
		return "", fmt.Errorf("failed to create task: %s", errorMsg)
	}

	utils.LogMessage(fmt.Sprintf("Task created with ID: %d", createTaskResp.TaskId), "process")

	getTaskData := map[string]interface{}{
		"clientKey": apiKey,
		"taskId":    createTaskResp.TaskId,
	}

	var result string
	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)

		jsonData, err = json.Marshal(getTaskData)
		if err != nil {
			continue
		}

		resp, err = http.Post(cs.twocaptchaApiUrl+"/getTaskResult", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			continue
		}

		var taskResult struct {
			ErrorId  int    `json:"errorId"`
			Status   string `json:"status"`
			Solution struct {
				Token string `json:"token"`
			} `json:"solution"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&taskResult); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		if taskResult.Status == "ready" {
			result = taskResult.Solution.Token
			utils.LogMessage("Captcha solved successfully!", "success")
			break
		}
	}

	if result == "" {
		return "", fmt.Errorf("failed to get captcha solution")
	}

	return result, nil
}
