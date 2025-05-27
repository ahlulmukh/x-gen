package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"
)

func LogMessage(message, messageType string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var colorPrinter *color.Color
	var symbol string

	switch messageType {
	case "info":
		colorPrinter = color.New(color.FgCyan)
		symbol = "[i]"
	case "success":
		colorPrinter = color.New(color.FgGreen)
		symbol = "[✓]"
	case "error":
		colorPrinter = color.New(color.FgRed)
		symbol = "[-]"
	case "warning":
		colorPrinter = color.New(color.FgYellow)
		symbol = "[!]"
	case "process":
		colorPrinter = color.New(color.FgHiCyan)
		symbol = "[>]"
	default:
		colorPrinter = color.New(color.Reset)
		symbol = "[*]"
	}

	logText := fmt.Sprintf("%s %s", symbol, message)
	fmt.Printf("[%s] ", timestamp)
	colorPrinter.Println(logText)
}

func GeneratePassword() string {
	rand.Seed(time.Now().UnixNano())

	firstLetter := string(rune(rand.Intn(26) + 65))

	otherLetters := make([]rune, 4)
	for i := range otherLetters {
		otherLetters[i] = rune(rand.Intn(26) + 97)
	}

	numbers := make([]rune, 3)
	for i := range numbers {
		numbers[i] = rune(rand.Intn(10) + 48)
	}

	return fmt.Sprintf("%s%s@%s!", firstLetter, string(otherLetters), string(numbers))
}

func GenerateEmail() string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	length := rand.Intn(5) + 8

	email := make([]byte, length)
	for i := range email {
		email[i] = charset[rand.Intn(len(charset))]
	}

	return string(email) + "@gmail.com"
}

func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

var (
	fileMutex sync.Mutex
)

var firstNames = []string{
	"Adi", "Budi", "Citra", "Dewi", "Eka", "Fajar", "Gita", "Hadi", "Indra", "Joko",
	"Kiki", "Lina", "Maya", "Nina", "Oki", "Putra", "Rina", "Sari", "Tono", "Wulan",
}

var lastNames = []string{
	"Pratama", "Santoso", "Wijaya", "Saputra", "Utami", "Rahmawati", "Putri", "Hidayat", "Susanto", "Permata",
	"Anggraini", "Setiawan", "Nugroho", "Saputri", "Wibowo", "Kurniawan", "Suryani", "Yuliana", "Ramadhan", "Fauzi",
}

func GenerateName() string {
	rand.Seed(time.Now().UnixNano())
	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func SaveAccountToFile(email, password string, username ...string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.OpenFile("accounts.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var accountLine string
	if len(username) > 0 && username[0] != "" {
		accountLine = fmt.Sprintf("%s:%s|%s:%s\n", email, password, username[0], password)
	} else {
		accountLine = fmt.Sprintf("%s:%s\n", email, password)
	}

	if _, err := file.WriteString(accountLine); err != nil {
		return err
	}

	return nil
}

func SaveTokenFile(email string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.OpenFile("authtoken.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	accountLine := fmt.Sprintf("%s\n", email)
	if _, err := file.WriteString(accountLine); err != nil {
		return err
	}

	return nil
}
