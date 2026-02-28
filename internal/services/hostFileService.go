package services

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

const marker string = "#4cus-guard"

func getPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

// func BlockURL(url string) {
// 	path := getPath()
// 	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
// 	if err != nil {
// 		log.Printf("Need Administrator rights: %v", err)
// 		return
// 	}
// 	defer file.Close()

// 	entry := "\n127.0.0.1 " + url + " " + marker
// 	if _, err := file.WriteString(entry); err != nil {
// 		log.Printf("Fail to Write file: %v", err)
// 	}
// }

// func UnblockURL(url string) {
// 	path := getPath()
// 	content, err := os.ReadFile(path)
// 	if err != nil {
// 		log.Printf("Fail to read file: %v", err)
// 		return
// 	}

// 	lines := strings.Split(string(content), "\n")
// 	var newLines []string
// 	for _, line := range lines {
// 		isDelete := strings.Contains(line, marker) && strings.Contains(line, url)
// 		if !isDelete {
// 			newLines = append(newLines, line)
// 		}
// 	}
// 	newContent := strings.Join(newLines, "\n")
// 	err = os.WriteFile(path, []byte(newContent), 0644)
// 	if err != nil {
// 		log.Printf("Need administrator rights: %v", err)
// 	}
// }

func BlockURL(rawUrl string) {
	domain := cleanDomain(rawUrl)
	path := getPath()

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Need Administrator rights: %v", err)
		return
	}
	defer file.Close()

	blockEntries := []string{
		fmt.Sprintf("127.0.0.1 %s %s", domain, marker),
		fmt.Sprintf("::1 %s %s", domain, marker),
		fmt.Sprintf("127.0.0.1 www.%s %s", domain, marker),
		fmt.Sprintf("::1 www.%s %s", domain, marker),
	}

	for _, entry := range blockEntries {
		if _, err := file.WriteString("\n" + entry); err != nil {
			log.Printf("Fail to Write file: %v", err)
		}
	}

	fmt.Printf("Blocked successfully: %s\n", domain)
	flushDNS()
}

func UnblockURL(rawUrl string) {
	domain := cleanDomain(rawUrl)
	path := getPath()

	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Fail to read file: %v", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		isDelete := strings.Contains(line, marker) && strings.Contains(line, domain)
		if !isDelete {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	newContent = strings.TrimRight(newContent, "\n") + "\n"

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("Need administrator rights: %v", err)
		return
	}

	fmt.Printf("Unblocked successfully %s\n", domain)
	flushDNS()
}
