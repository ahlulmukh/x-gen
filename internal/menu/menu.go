package menu

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"x-gen/internal/utils"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

type MenuHandler struct {
	reader *bufio.Reader
}

func NewMenuHandler() *MenuHandler {
	return &MenuHandler{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (m *MenuHandler) ShowMainMenu(version string) {
	m.showBanner(version)

	if !m.configExists() {
		m.createConfig()
	}

	for {
		choice := m.showMenuOptions()
		switch choice {
		case "1":
			m.RunTwitterGenerator()
			m.showBanner(version)
		case "2":
			m.EditConfig()
			m.showBanner(version)
		case "3":
			m.ShowFileInfo()
			m.showBanner(version)
		case "4":
			os.Exit(0)
		default:
			utils.LogMessage("Invalid choice", "error")
		}
	}
}

func (m *MenuHandler) showBanner(version string) {
	//utils.ClearScreen()
	myFigure := figure.NewFigure("Twitter Gen", "", true)
	figureStr := myFigure.String()
	fmt.Println(color.HiYellowString(figureStr))

	fmt.Println(color.HiBlueString("🔹 Made by : ") + color.WhiteString("Ahlul Mukhramin") + "  |  " +
		color.HiBlueString("📦 GitHub: ") + color.WhiteString("github.com/ahlulmukh") + "  |  " +
		color.HiBlueString("✅ Version: ") + color.WhiteString(version))
}

func (m *MenuHandler) showMenuOptions() string {
	fmt.Println(color.CyanString("\nMain Menu:"))
	fmt.Println(color.HiGreenString("1. X.com Account Generator"))
	fmt.Println(color.HiBlueString("2. Edit Config"))
	fmt.Println(color.MagentaString("3. Information"))
	fmt.Println(color.RedString("4. Exit"))
	fmt.Print(color.HiCyanString("Enter your choice (1-4): "))

	choice, _ := m.reader.ReadString('\n')
	return strings.TrimSpace(choice)
}

func (m *MenuHandler) ShowFileInfo() {
	//utils.ClearScreen()

	fmt.Println(color.YellowString("3. proxy.txt (optional)"))
	fmt.Println("   - Format: user:pass@host:port")
	fmt.Println("   - Example: http://puqus:gaming@yesyes.com:8080")
	fmt.Println()

	m.waitForEnter()
}

func (m *MenuHandler) waitForEnter() {
	fmt.Print("Press Enter to continue...")
	m.reader.ReadBytes('\n')
}
