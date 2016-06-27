package main

import (
	"flag"
	"fmt"
	slog "log"
	apiService "longChat-Server/apiService/service"
	"longChat-Server/common/util/config"
	idService "longChat-Server/idService/service"
)

func main() {
	pconfig := flag.String("config", "../config.cfg", "config file")
	psection := flag.String("section", "dev", "section of config file to apply")
	flag.Parse()
	env := map[string]string{
		"config":  *pconfig,
		"section": *psection,
	}
	config.InitConfigEnv(env)
	err := config.LoadConfigFile()
	if err != nil {
		slog.Fatalf("LoadConfigFile from %s failed. err=%v\n", *pconfig, err)
	}

	idService.InitIdService(true)
	id, err := apiService.CreateUser()
	fmt.Println(id, err)
}
