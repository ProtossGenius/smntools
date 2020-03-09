package smcfg

import (
	"os/user"
)

func GetUserHome() string {
	user, err := user.Current()
	if err == nil {
		return user.HomeDir
	}
	return "/"

}

var userHome = GetUserHome()

func GetCfgPath() string {
	return userHome + ".smcfg/"
}
