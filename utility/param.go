package utility

import (
	"strings"

	"github.com/gogf/gf/v2/os/genv"
)

func GetEnvParam(name string) string {
	v := genv.GetWithCmd(name)
	if v != nil {
		val := v.String()
		if val != "" {
			return val
		}
	}

	// Allow platform-friendly variables such as DATABASE_LINK for database.link.
	alt := strings.ToUpper(strings.ReplaceAll(name, ".", "_"))
	v = genv.GetWithCmd(alt)
	if v != nil {
		return v.String()
	}

	return ""
}
