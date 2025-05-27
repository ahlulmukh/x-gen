package menu

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"x-gen/internal/proxy"
	"x-gen/internal/utils"
	"x-gen/internal/xgen"
)

func (m *MenuHandler) RunTwitterGenerator() {

	count := m.getInputGenerator("How many accounts? ", 1)
	threads := m.getInputGenerator("Threads count? ", 1)

	m.startGeneratorAccount(count, threads)
}

func (m *MenuHandler) getInputGenerator(prompt string, min int) int {
	fmt.Print(prompt)
	input, _ := m.reader.ReadString('\n')
	val, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || val < min {
		utils.LogMessage("Invalid input", "error")
		return m.getInputGenerator(prompt, min)
	}
	return val
}

func (m *MenuHandler) startGeneratorAccount(count, threads int) {
	proxy.LoadProxies()

	var wg sync.WaitGroup
	jobs := make(chan int, count)
	successCh := make(chan int, count)

	for w := 0; w < threads; w++ {
		wg.Add(1)
		go m.generateWorkerX(w, &wg, jobs, successCh, count)
	}

	for i := 0; i < count; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	m.showResultsGen(successCh, count)
}

func (m *MenuHandler) generateWorkerX(_ int, wg *sync.WaitGroup, jobs <-chan int,
	successCh chan<- int, _ int) {
	defer wg.Done()

	for idx := range jobs {
		proxy, err := proxy.GetRandomProxy()
		if err != nil {
			utils.LogMessage(fmt.Sprintf("Failed to get proxy for job %d: %v", idx+1, err), "error")
			continue
		}
		slx := xgen.NewXGenerator(proxy)

		if _, err := slx.GenerateXAccount(); err == nil {
			successCh <- 1
		}
	}
}

func (m *MenuHandler) showResultsGen(successCh chan int, total int) {
	close(successCh)
	success := len(successCh)

	utils.LogMessage("Process completed!", "success")
	utils.LogMessage(
		fmt.Sprintf("Success: %d/%d", success, total), "info")
	m.waitForEnter()
}
