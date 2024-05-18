package helpers

import (
	"fmt"
	"strings"
)

func FormatTimeFromUnix(timestamp int) string {
	return fmt.Sprintf("%02d:%02d:%02d",
		(timestamp-(timestamp%3600))/3600,
		(timestamp-(timestamp%60)-(timestamp-(timestamp%3600)))/60,
		timestamp%60,
	)
}

func GetFormattedDomain(domain string) string {
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return domain
	} else {
		return "https://" + domain
	}
}
