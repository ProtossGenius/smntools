package smcfg

import (
	"os/user"
)

func GetCfgPath() (string, error) {
	user, err := user.Current()
	if err == nil {
		return user.HomeDir + "/.smcfg/", nil
	}
	return "", err
}
