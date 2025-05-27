# X-Gen - Automated X.com (Twitter) Account Generator

🤖 **Powerful automated tool for creating X.com accounts with email verification and captcha solving**

![Version](https://img.shields.io/badge/version-0.0.1-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24.2-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

![Preview](preview.gif)

## 📋 Table of Contents

- [Features](#-features)
- [Project Structure](#-project-structure)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [File Formats](#-file-formats)
- [Email Service Configuration](#-email-service-configuration)
- [Captcha Services](#-captcha-services)
- [Proxy Support](#-proxy-support)
- [Output Files](#-output-files)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [Disclaimer](#-disclaimer)

## 🚀 Features

- ✅ **Automated X.com Account Creation** - Fully automated account registration process
- 📧 **Temporary Email Integration** - Uses Mail.tm for automatic email verification
- 🔐 **Multi-Captcha Support** - AntiCaptcha and 2Captcha integration
- 🌐 **Proxy Support** - HTTP/SOCKS proxy rotation for anonymity
- 📊 **Multi-threading** - Concurrent account generation for efficiency
- 🎯 **Indonesian Name Generation** - Built-in Indonesian name generator
- 💾 **Auto-Save Results** - Automatically saves accounts and auth tokens
- 🔄 **Auto-Update System** - Built-in update checker and installer
- 🎨 **Beautiful CLI Interface** - Colorful and user-friendly interface
- ⚙️ **Configuration Management** - Easy config setup and editing

## 📁 Project Structure

```
x-gen/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── captcha/
│   │   └── captcha.go            # Captcha solving services
│   ├── menu/
│   │   ├── config.go             # Configuration management
│   │   ├── menu.go               # Main menu interface
│   │   └── x-gen.go              # Account generation workflow
│   ├── proxy/
│   │   └── proxy.go              # Proxy management and rotation
│   ├── updater/
│   │   └── updater.go            # Auto-update functionality
│   ├── utils/
│   │   ├── mailhandler.go        # Mail.tm email handling
│   │   └── utils.go              # Utility functions and logging
│   └── xgen/
│       ├── x-account.go          # Core X.com account creation logic
│       └── xgen.go               # X account generator struct
├── accounts.txt                   # Generated accounts (email:password:username)
├── authtoken.txt                  # Generated auth tokens
├── config.json                    # Application configuration
├── proxy.txt                      # Proxy list (optional)
├── go.mod                         # Go module dependencies
├── go.sum                         # Go module checksums
└── README.md                      # Project documentation
```

## 🔧 Installation

### Prerequisites

- Go 1.24.2 or higher
- Internet connection
- Valid captcha service API key (AntiCaptcha or 2Captcha)

### Steps

1. **Clone or download the project**
   ```bash
   git clone <repository-url>
   cd x-gen
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Build the application**
   ```bash
   go build -o x-gen.exe cmd/main.go
   ```

4. **Run the application**
   ```bash
   ./x-gen.exe
   ```

## ⚙️ Configuration

On first run, the application will guide you through creating a configuration file.

### Setting up Captcha Service

Choose one of the supported captcha services:

#### Option 1: AntiCaptcha
```json
{
  "captchaServices": {
    "captchaUsing": "antiCaptcha",
    "antiCaptchaApikey": ["your-anticaptcha-api-key"]
  }
}
```

#### Option 2: 2Captcha
```json
{
  "captchaServices": {
    "captchaUsing": "2captcha",
    "captcha2Apikey": ["your-2captcha-api-key"]
  }
}
```

### Editing Configuration

You can modify the configuration anytime by:
1. Selecting option `2. Edit Config` from the main menu
2. Manually editing `config.json` file

## 🎯 Usage

### Basic Usage

1. **Start the application**
   ```bash
   ./x-gen.exe
   ```

2. **Main Menu Options**
   ```
   1. X.com Account Generator  - Generate new accounts
   2. Edit Config             - Modify configuration
   3. Information            - View file format information
   4. Exit                   - Close application
   ```

3. **Generate Accounts**
   - Select option 1
   - Enter number of accounts to generate
   - Enter number of threads (concurrent workers)
   - Wait for completion

### Advanced Usage

- **Multi-threading**: Use 2-5 threads for optimal performance
- **Proxy Rotation**: Add proxies to `proxy.txt` for better success rate
- **Batch Generation**: Generate multiple accounts in one session

## 📄 File Formats

### accounts.txt
Generated accounts are saved in the format:
```
email@mail.tm:generatedPassword:@username
email2@mail.tm:generatedPassword2:@username2
```

### authtoken.txt
Authentication tokens are saved one per line:
```
auth_token_1
auth_token_2
auth_token_3
```

### proxy.txt (optional)
Proxy list in the format:
```
user:pass@host:port
http://username:password@proxy.example.com:8080
socks5://user:pass@proxy.example.com:1080
```

## 📧 Email Service Configuration

### Mail.tm Integration

The application uses **Mail.tm** service for temporary email creation:

- **Automatic Account Creation**: Creates temporary email accounts automatically
- **Email Verification**: Automatically retrieves verification codes
- **No Manual Setup**: No additional configuration required
- **Temporary Emails**: Emails are temporary and will be deleted automatically

### How It Works

1. Creates a temporary email account using Mail.tm API
2. Uses the email for X.com registration
3. Automatically polls for verification emails
4. Extracts verification code using regex patterns
5. Completes the verification process

### Fallback Method

If automatic email verification fails, the application will:
- Display the email address used
- Prompt for manual verification code entry
- Continue with the registration process

## 🔐 Captcha Services

### Supported Services

#### 1. AntiCaptcha
- **Website**: https://anti-captcha.com/
- **Pricing**: Pay per captcha solved
- **Setup**: Get API key from dashboard
- **Recommended**: Yes, generally faster

#### 2. 2Captcha
- **Website**: https://2captcha.com/
- **Pricing**: Pay per captcha solved
- **Setup**: Get API key from account settings
- **Recommended**: Good alternative option

### Captcha Service Configuration

1. **Sign up** for your preferred captcha service
2. **Get API key** from your account dashboard
3. **Add funds** to your account
4. **Configure** the API key in the application

## 🌐 Proxy Support

### Supported Proxy Types

- **HTTP Proxies**: `http://user:pass@host:port`
- **HTTPS Proxies**: `https://user:pass@host:port`
- **SOCKS5 Proxies**: `socks5://user:pass@host:port`

### Proxy Configuration

1. **Create** `proxy.txt` file in the project root
2. **Add proxies** one per line in the format shown above
3. **No authentication** proxies: `host:port`
4. **With authentication**: `user:pass@host:port`

### Proxy Features

- **Automatic Rotation**: Randomly selects proxies for each request
- **IP Checking**: Verifies IP changes between requests
- **Error Handling**: Automatically switches proxies on failure
- **Optional Usage**: Application works without proxies

## 📊 Output Files

### Generated Files

1. **accounts.txt**
   - Contains all successfully created accounts
   - Format: `email:password:username`
   - Automatically updated after each successful generation

2. **authtoken.txt**
   - Contains authentication tokens for created accounts
   - One token per line
   - Can be used for API access

3. **config.json**
   - Application configuration
   - Captcha service settings
   - Can be edited manually or through the interface

### File Location

All output files are created in the same directory as the executable.

## 🔧 Troubleshooting

### Common Issues

#### 1. Captcha Solving Fails
- **Check API key**: Ensure your captcha service API key is correct
- **Check balance**: Verify you have sufficient funds in your captcha service account
- **Service status**: Check if the captcha service is operational

#### 2. Email Verification Fails
- **Network issues**: Check your internet connection
- **Mail.tm status**: The service might be temporarily unavailable
- **Manual verification**: Use the fallback manual verification option

#### 3. Account Creation Fails
- **Rate limiting**: X.com might be rate limiting requests
- **Proxy issues**: Try using different proxies or no proxy
- **Captcha issues**: Ensure captcha service is working properly

#### 4. Proxy Connection Issues
- **Format check**: Verify proxy format is correct
- **Credentials**: Ensure proxy username/password are correct
- **Proxy status**: Test if proxies are working

### Debug Tips

1. **Check logs**: Application provides detailed logging with timestamps
2. **Start small**: Test with 1 account and 1 thread first
3. **Monitor success rate**: Check the success/failure ratio
4. **Update regularly**: Keep the application updated

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## ⚠️ Disclaimer

This tool is for educational and testing purposes only. Users are responsible for:

- **Compliance**: Ensuring compliance with X.com's Terms of Service
- **Legal use**: Using the tool within legal boundaries
- **Ethical use**: Not using for spam or malicious activities
- **Rate limits**: Respecting platform rate limits and policies

The developers are not responsible for any misuse of this tool or any consequences arising from its use.

**⭐ If this project helped you, please give it a star!**