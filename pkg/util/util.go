package util

import "fmt"

// GetWebsiteURL returs the URL for accessing the website
func GetWebsiteURL(ip string) string {
	return fmt.Sprintf("http://%s:1313", ip)
}
