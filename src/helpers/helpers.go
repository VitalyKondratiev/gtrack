package helpers

import "fmt"

func FormatTimeFromUnix(timestamp int) string {
	return fmt.Sprintf("%02d:%02d:%02d",
		(timestamp-(timestamp%3600))/3600,
		(timestamp-(timestamp%60)-(timestamp-(timestamp%3600)))/60,
		timestamp%60,
	)
}
