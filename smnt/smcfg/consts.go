package smcfg

import "runtime"

func init() {
	switch runtime.GOOS {
	case "windows":
		cfgPath = "C:/.smcfg/"
	default:
		cfgPath = "~/.smcfg/"
	}
}

var (
	cfgPath string
)

func GetCfgPath() string {
	return cfgPath
}
