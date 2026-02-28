package services

import (
	"fmt"
	"os/exec"
	"runtime"
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
	var candidates [][]string

	switch runtime.GOOS {
	case "windows":
		candidates = [][]string{
			{"ipconfig", "/flushdns"},
		}
	case "linux":
		// Try resolvers in order of prevalence.
		// Only the one that is actually installed will succeed.
		candidates = [][]string{
			{"resolvectl", "flush-caches"},          // systemd-resolved (Ubuntu 20.04+)
			{"systemd-resolve", "--flush-caches"},   // older systemd-resolved
			{"nscd", "-i", "hosts"},                 // nscd
			{"dnsmasq", "--clear-on-reload"},        // dnsmasq
		}
	default:
		return
	}

	for _, args := range candidates {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err == nil {
			return
		}
	}

	fmt.Println("Can't Flush DNS: no supported resolver found")
}