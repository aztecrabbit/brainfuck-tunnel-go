package main

import (
	"os"
	"fmt"
	"time"
	"strconv"

	"github.com/aztecrabbit/liblog"
	"github.com/aztecrabbit/libutils"
	"github.com/aztecrabbit/libinject"
	"github.com/aztecrabbit/libredsocks"
	"github.com/aztecrabbit/libproxyrotator"
	"github.com/aztecrabbit/brainfuck-tunnel-go/src/libsshclient"
)

const (
	appName = "Brainfuck Tunnel"
	appVersionName = "Go"
	appVersionCode = "200120"

	copyrightYear = "2020"
	copyrightAuthor = "Aztec Rabbit"
)

var (
	InterruptHandler = new(libutils.InterruptHandler)
	Redsocks = new(libredsocks.Redsocks)
)

type Config struct {
	ProxyRotator *libproxyrotator.Config
	Inject *libinject.Config
	SshClientThreads int
	SshClient *libsshclient.Config
}

func init() {
	InterruptHandler.Handle = func() {
		libredsocks.Stop(Redsocks)
		libsshclient.Stop()
		liblog.LogKeyboardInterrupt()
	}
	InterruptHandler.Start()
}

func main() {
	liblog.Header(
		[]string{
			fmt.Sprintf("%s [%s Version. %s]", appName, appVersionName, appVersionCode),
			fmt.Sprintf("(c) %s %s.", copyrightYear, copyrightAuthor),
		},
		liblog.Colors["G1"],
	)

	config := new(Config)
	defaultConfig := new(Config)
	defaultConfig.ProxyRotator = libproxyrotator.DefaultConfig
	defaultConfig.Inject = libinject.DefaultConfig
	defaultConfig.SshClientThreads = 4
	defaultConfig.SshClient = libsshclient.DefaultConfig

	libutils.JsonReadWrite(libutils.RealPath("config.json"), config, defaultConfig)

	ProxyRotator := new(libproxyrotator.ProxyRotator)
	ProxyRotator.Config = config.ProxyRotator

	Inject := new(libinject.Inject)
	Inject.Redsocks = Redsocks
	Inject.Config = config.Inject

	if len(os.Args) > 1 {
		Inject.Config.Port = os.Args[1]
	}

	go ProxyRotator.Start()
	go Inject.Start()

	time.Sleep(200 * time.Millisecond)

	liblog.LogInfo("Proxy Rotator running on port " + ProxyRotator.Config.Port, "INFO", liblog.Colors["G1"])
	liblog.LogInfo("Inject running on port " + Inject.Config.Port, "INFO", liblog.Colors["G1"])

	Redsocks.Config = libredsocks.DefaultConfig
	Redsocks.Start()

	for i := 1; i <= config.SshClientThreads; i++ {
		SshClient := new(libsshclient.SshClient)
		SshClient.ProxyRotator = ProxyRotator
		SshClient.Config = config.SshClient
		SshClient.InjectPort = Inject.Config.Port
		SshClient.ListenPort = strconv.Itoa(libutils.Atoi(ProxyRotator.Config.Port) + i)
		SshClient.Verbose = false
		SshClient.Loop = true

		go SshClient.Start()
	}

	InterruptHandler.Wait()
}
