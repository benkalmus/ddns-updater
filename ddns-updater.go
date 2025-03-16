package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// GetPublicIP fetches the public IP from an external API.
// will attempt to fetch IP from another source, if one fails
func GetPublicIP() (string, error) {
	backups := []string{
		"https://api.ipify.org",
		"http://checkip.amazonaws.com",
	}
	for _, endpoint := range backups {
		resp, err := http.Get(endpoint)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		ip, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		ipStr := removeAllWhitespace(string(ip))
		return ipStr, nil
	}

	return "", fmt.Errorf("all get public IP endpoints failed")
}

func removeAllWhitespace(input string) string {
	result := strings.ReplaceAll(input, " ", "")
	result = strings.ReplaceAll(result, "\n", "")
	result = strings.ReplaceAll(result, "\r", "")

	return result
}

// ReadLastIP reads the last known IP from a file.
func ReadLastIP() string {
	data, err := os.ReadFile("last_ip.txt")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// WriteLastIP stores the current public IP to a file.
func WriteLastIP(ip string) error {
	return os.WriteFile("last_ip.txt", []byte(ip), 0644)
}

// GetEnvHash computes an MD5 hash of the .env file
func GetEnvHash() (string, error) {
	data, err := os.ReadFile(".env")
	if err != nil {
		return "", err
	}
	hash := fmt.Sprintf("%x", md5.Sum(data))
	return hash, nil
}

// ReadLastEnvHash gets the stored .env hash
func ReadLastEnvHash() string {
	data, err := os.ReadFile("last_env_hash.txt")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// WriteLastEnvHash saves the .env hash
func WriteLastEnvHash(hash string) error {
	return os.WriteFile("last_env_hash.txt", []byte(hash), 0644)
}

// UpdateDNS updates the DNS entry for a specific host.
func UpdateDNS(host, domain, password, ip string) error {
	urlStr := fmt.Sprintf("https://dynamicdns.park-your-domain.com/update?host=%s&domain=%s&password=%s",
		url.QueryEscape(host), url.QueryEscape(domain), url.QueryEscape(password))
	// dynamicdns will use our hosts's IP if ip query parameter is not set
	if ip != "" {
		urlStr = fmt.Sprintf("%s&ip=%s", urlStr, url.QueryEscape(ip))
	}

	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Updated %s.%s -> %s\nResponse: %s\n", host, domain, ip, string(body))

	return nil
}

func main() {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	domain := os.Getenv("DOMAIN_NAME")
	password := os.Getenv("DDNS_PASSWORD")
	hosts := strings.Split(os.Getenv("HOSTS"), ",") // Multiple hosts

	// Get current .env hash
	currentEnvHash, err := GetEnvHash()
	if err != nil {
		log.Fatalf("Error computing .env hash: %v", err)
	}

	// Read the last stored .env hash
	lastEnvHash := ReadLastEnvHash()

	// Check if .env file changed
	envChanged := currentEnvHash != lastEnvHash

	// Get the public IP
	publicIP, err := GetPublicIP()
	if err != nil {
		log.Printf("Error fetching public IP: %v", err)
	}

	// Read last known IP to avoid unnecessary updates
	lastIP := ReadLastIP()
	if publicIP == lastIP && !envChanged {
		fmt.Printf("IP changed: %v, .env changed: %v, no update needed.\n", publicIP != lastIP, envChanged)
		return
	}

	// Update DNS for each host
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if err := UpdateDNS(host, domain, password, publicIP); err != nil {
			log.Printf("Failed to update DNS for %s: %v", host, err)
		}
	}

	// Store the updated IP
	if err := WriteLastIP(publicIP); err != nil {
		log.Printf("Failed to write last IP: %v", err)
	}

	// store env hash, so that if the env file changes we will update DNS
	if err := WriteLastEnvHash(currentEnvHash); err != nil {
		log.Printf("Failed to write last .env hash: %v", err)
	}
}
