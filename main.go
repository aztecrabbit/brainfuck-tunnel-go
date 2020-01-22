package main

import (
	"os"
	"fmt"
	"time"
	"sync"
	"strconv"

	"github.com/aztecrabbit/liblog"
	"github.com/aztecrabbit/libutils"
	"github.com/aztecrabbit/libinject"
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

type Config struct {
	ProxyRotator *libproxyrotator.Config
	Inject *libinject.Config
	SshClient *libsshclient.Config
	SshClientThreads int
}

func init() {
	InterruptHandler := &libutils.InterruptHandler{
		Handle: func() {
			libsshclient.Stop()
			liblog.LogKeyboardInterrupt()
		},
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
	configDefault := new(Config)
	configDefault.SshClientThreads = 4

	// Proxy Rotator config
	configDefault.ProxyRotator = libproxyrotator.ConfigDefault

	// Inject config
	configDefault.Inject = libinject.ConfigDefault

	// Ssh Client config
	configDefault.SshClient = libsshclient.ConfigDefault
	configDefault.SshClient.ProxyPort = configDefault.Inject.Port

	libutils.JsonReadWrite(libutils.RealPath("config.json"), config, configDefault)

	ProxyRotator := new(libproxyrotator.ProxyRotator)
	ProxyRotator.Config = config.ProxyRotator

	Inject := new(libinject.Inject)
	Inject.Config = config.Inject

	if len(os.Args) > 1 {
		Inject.Config.Port = os.Args[1]
	}

	channel := make(chan bool, config.SshClientThreads)

	var wg sync.WaitGroup

	for i := 1; i <= config.SshClientThreads; i++ {
		wg.Add(1)

		SshClient := new(libsshclient.SshClient)
		SshClient.Config = config.SshClient
		SshClient.Config.ProxyPort = Inject.Config.Port
		SshClient.ListenPort = strconv.Itoa(libutils.Atoi(ProxyRotator.Config.Port) + i)
		SshClient.Verbose = false
		SshClient.Loop = true

		ProxyRotator.Proxies = append(ProxyRotator.Proxies, "0.0.0.0:" + SshClient.ListenPort)

		go SshClient.Start(&wg, channel)
	}

	go ProxyRotator.Start()
	go Inject.Start()

	time.Sleep(200 * time.Millisecond)

	liblog.LogInfo("Proxy Rotator running on port " + ProxyRotator.Config.Port, "INFO", liblog.Colors["G1"])
	liblog.LogInfo("Inject running on port " + Inject.Config.Port, "INFO", liblog.Colors["G1"])

	for i := 0; i < config.SshClientThreads; i++ {
		channel <- true
	}

	wg.Wait()
}
