package services

import (
	"fmt"
	"os/exec"
	"strings"
)

func cleanDomain(input string) string {
	domain := strings.TrimSpace(input)
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.Split(domain, "/")[0] 
	domain = strings.TrimPrefix(domain, "www.")
	return domain
}

func flushDNS() {
	cmd := exec.Command("ipconfig", "/flushdns")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Can't Flush DNS: %v\n", err)
	}
}