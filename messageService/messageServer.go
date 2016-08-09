package main

import (
	"flag"
	slog "log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/messageService/message"
	"github.com/longchat/longChat-Server/storageService/storage"
)

var cpuf *os.File

func main() {
	pconfig := flag.String("config", "../config.cfg", "config file")
	psection := flag.String("section", "dev", "section of config file to apply")
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile := flag.String("memprofile", "", "write mem profile to file")

	flag.Parse()
	if *cpuProfile != "" {
		var err error
		cpuf, err = os.Create(*cpuProfile)
		if err != nil {
			slog.Fatal(err)
		}
		pprof.StartCPUProfile(cpuf)
	}
	go func() {
		if *memProfile != "" || *cpuProfile != "" {
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			for {
				s := <-c
				if s.(syscall.Signal) == syscall.SIGINT || s.(syscall.Signal) == syscall.SIGTERM {
					break
				}
			}

			if *memProfile != "" {
				memf, err := os.Create(*memProfile)
				if err != nil {
					slog.Fatal(err)
				}
				pprof.WriteHeapProfile(memf)
				memf.Close()
			}
			if *cpuProfile != "" {
				pprof.StopCPUProfile()
				cpuf.Close()
			}
		}
		os.Exit(0)
	}()

	config.InitConfig(pconfig, psection)

	accPath, err := config.GetConfigString(consts.AccessLogPath)
	if err != nil {
		slog.Fatalf(consts.ErrGetConfigFailed(consts.AccessLogPath, err))
	}
	errPath, err := config.GetConfigString(consts.ErrorLogPath)
	if err != nil {
		slog.Fatalf(consts.ErrGetConfigFailed(consts.ErrorLogPath, err))
	}
	err = log.InitLogger(errPath, accPath, 1024, 5*1024)
	if err != nil {
		slog.Fatalf("init log failed!err:=%v\n", err)
	}
	defer log.FiniLogger()
	addr, err := config.GetConfigString(consts.MsgServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.MsgServiceAddress, err))
	}
	parentAddr, err := config.GetConfigString(consts.ParentServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.ParentServiceAddress, err))
	}
	isLeaf, err := config.GetConfigBool(consts.IsLeafServer)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.IsLeafServer, err))
	}
	store, err := storage.NewStorage()
	if err != nil {
		slog.Fatalln("init store failed!", err)
	}
	defer store.Close()
	if err != nil {
		slog.Fatalf("init DB failed!err:=%v\n", err)
	}
	message.StartServer(store, addr, parentAddr, isLeaf)
}
