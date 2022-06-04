package settingsfamily

import (
	"doc-management/internal/hashing"
	"strings"
)

func GetAddress(settingName string) string {
	addr := "000000"
	parts := strings.Split(settingName, ".")
	for i := 0; i < 4; i++ {
		if i < len(parts) {
			addr += hashing.CalculateSHA256(parts[i])[:16]
		} else {
			addr += hashing.CalculateSHA256("")[:16]
		}
	}
	return addr
}
