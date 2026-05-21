package i18n

import (
	"os"
	"strings"
)

var lang string

func init() {
	lang = detectLang()
}

func detectLang() string {
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		val := os.Getenv(env)
		if strings.Contains(strings.ToLower(val), "zh") {
			return "zh"
		}
	}
	return "en"
}

// T returns localized text. zh for Chinese, en for everything else.
func T(zh, en string) string {
	if lang == "zh" {
		return zh
	}
	return en
}
