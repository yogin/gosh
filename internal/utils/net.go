package utils

func SafeIP(ip *string) string {
	if ip == nil {
		return ""
	}

	return *ip
}
