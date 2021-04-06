package core

import "os"

func ReadConfigPath(confStr string) string {
	conf := ""
	if confStr == "" {
		tmp, _ := GetExecPath()
		conf = tmp + "al.conf"
	} else {
		conf, _ = ParsePath(confStr)
	}
	Log.Info("config path=", conf)

	_, err := os.Stat(conf)
	if err != nil {
		if os.IsNotExist(err) {
			Errorln("config file", conf+" not exists")
			Log.Error("config file not exists")
		} else {
			Errorln("unknown error", err)
			Log.Error("unknown error", err)
		}
		os.Exit(0)
	}

	return conf
}
