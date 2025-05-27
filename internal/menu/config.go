package menu

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"solix-bot/internal/captcha"
	"solix-bot/internal/utils"
	"strings"

	"github.com/fatih/color"
)

func (m *MenuHandler) configExists() bool {
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		return false
	}
	return true
}

func (m *MenuHandler) createConfig() {
	utils.LogMessage("Creating new config...", "info")

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(color.CyanString("\nSelect your captcha service:"))
	fmt.Println(color.YellowString("1. AntiCaptcha"))
	fmt.Println(color.BlueString("2. 2Captcha"))
	// fmt.Println(color.WhiteString("3. Private"))
	fmt.Print(color.CyanString("Enter your choice (1-2): "))

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var captchaUsing string
	var urlPrivate string
	var antiCaptchaApikey []string
	var captcha2Apikey []string

	switch choice {
	case "1":
		captchaUsing = "antiCaptcha"
		fmt.Print(color.CyanString("Enter your AntiCaptcha API key: "))
		apiKey, _ := reader.ReadString('\n')
		antiCaptchaApikey = []string{strings.TrimSpace(apiKey)}
	case "2":
		captchaUsing = "2captcha"
		fmt.Print(color.CyanString("Enter your 2Captcha API key: "))
		apiKey, _ := reader.ReadString('\n')
		captcha2Apikey = []string{strings.TrimSpace(apiKey)}
	// case "3":
	// 	captchaUsing = "private"
	// 	fmt.Print("Enter your private captcha solver URL: ")
	// 	url, _ := reader.ReadString('\n')
	// 	urlPrivate = strings.TrimSpace(url)
	default:
		utils.LogMessage("Invalid choice, using AntiCaptcha by default", "warning")
		captchaUsing = "antiCaptcha"
		fmt.Print(color.CyanString("Enter your AntiCaptcha API key: "))
		apiKey, _ := reader.ReadString('\n')
		antiCaptchaApikey = []string{strings.TrimSpace(apiKey)}
	}

	config := captcha.Config{
		CaptchaServices: struct {
			CaptchaUsing      string   `json:"captchaUsing"`
			UrlPrivate        string   `json:"urlPrivate"`
			AntiCaptchaApikey []string `json:"antiCaptchaApikey"`
			Captcha2Apikey    []string `json:"captcha2Apikey"`
		}{
			CaptchaUsing:      captchaUsing,
			UrlPrivate:        urlPrivate,
			AntiCaptchaApikey: antiCaptchaApikey,
			Captcha2Apikey:    captcha2Apikey,
		},
	}

	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		utils.LogMessage("Failed to create config: "+err.Error(), "error")
		return
	}

	err = os.WriteFile("config.json", file, 0644)
	if err != nil {
		utils.LogMessage("Failed to save config: "+err.Error(), "error")
		return
	}

	utils.LogMessage("Config file created successfully!", "success")
}

func (m *MenuHandler) EditConfig() {
	config := captcha.LoadConfig()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(color.CyanString("\nCurrent Config:"))
	fmt.Printf(color.GreenString("1. Captcha Service: %s\n"), config.CaptchaServices.CaptchaUsing)
	if config.CaptchaServices.CaptchaUsing == "antiCaptcha" {
		fmt.Printf(color.RedString("2. AntiCaptcha API Key: %s\n"), config.CaptchaServices.AntiCaptchaApikey[0])
	} else if config.CaptchaServices.CaptchaUsing == "2captcha" {
		fmt.Printf(color.CyanString("2. 2Captcha API Key: %s\n"), config.CaptchaServices.Captcha2Apikey[0])
	} else {
		fmt.Printf("2. Private Captcha URL: %s\n", config.CaptchaServices.UrlPrivate)
	}

	fmt.Println(color.YellowString("\nWhat would you like to change?"))
	fmt.Println(color.GreenString("1. Change captcha service"))
	fmt.Println(color.RedString("2. Change API key"))
	fmt.Println(color.BlueString("3. Back to main menu"))
	fmt.Print(color.YellowString("Enter your choice (1-3): "))

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Println("\nSelect your captcha service:")
		fmt.Println("1. AntiCaptcha")
		fmt.Println("2. 2Captcha")
		// fmt.Println("3. Private")
		fmt.Print("Enter your choice (1-2): ")

		newChoice, _ := reader.ReadString('\n')
		newChoice = strings.TrimSpace(newChoice)

		switch newChoice {
		case "1":
			config.CaptchaServices.CaptchaUsing = "antiCaptcha"
			fmt.Print("Enter your AntiCaptcha API key: ")
			apiKey, _ := reader.ReadString('\n')
			config.CaptchaServices.AntiCaptchaApikey = []string{strings.TrimSpace(apiKey)}
		case "2":
			config.CaptchaServices.CaptchaUsing = "2captcha"
			fmt.Print("Enter your 2Captcha API key: ")
			apiKey, _ := reader.ReadString('\n')
			config.CaptchaServices.Captcha2Apikey = []string{strings.TrimSpace(apiKey)}
		// case "3":
		// 	config.CaptchaServices.CaptchaUsing = "private"
		// 	fmt.Print("Enter your private captcha solver URL: ")
		// 	url, _ := reader.ReadString('\n')
		// 	config.CaptchaServices.UrlPrivate = strings.TrimSpace(url)
		default:
			utils.LogMessage("Invalid choice, no changes made", "error")
			return
		}
	case "2":
		if config.CaptchaServices.CaptchaUsing == "antiCaptcha" {
			fmt.Print("Enter new AntiCaptcha API key: ")
			apiKey, _ := reader.ReadString('\n')
			config.CaptchaServices.AntiCaptchaApikey = []string{strings.TrimSpace(apiKey)}
		} else if config.CaptchaServices.CaptchaUsing == "2captcha" {
			fmt.Print("Enter new 2Captcha API key: ")
			apiKey, _ := reader.ReadString('\n')
			config.CaptchaServices.Captcha2Apikey = []string{strings.TrimSpace(apiKey)}
		} else {
			fmt.Print("Enter new private captcha solver URL: ")
			url, _ := reader.ReadString('\n')
			config.CaptchaServices.UrlPrivate = strings.TrimSpace(url)
		}
	case "3":
		return
	default:
		utils.LogMessage("Invalid choice, no changes made", "error")
		return
	}

	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		utils.LogMessage("Failed to update config: "+err.Error(), "error")
		return
	}

	err = os.WriteFile("config.json", file, 0644)
	if err != nil {
		utils.LogMessage("Failed to save config: "+err.Error(), "error")
		return
	}

	utils.LogMessage("Config updated successfully!", "success")

	m.saveConfig(config)
}

func (m *MenuHandler) saveConfig(config *captcha.Config) error {
	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("config.json", file, 0644)
}
