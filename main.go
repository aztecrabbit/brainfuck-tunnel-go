package main

import (
	"os"
	"os/signal"
	"fmt"
	"time"
	"sync"
	"syscall"
	"strconv"

	"github.com/aztecrabbit/liblog"
	"github.com/aztecrabbit/libinject"
	"github.com/aztecrabbit/libproxyrotator"
	"github.com/aztecrabbit/brainfuck-tunnel-go/src/sshclient"
)

const (
	appName = "Brainfuck Tunnel"
	appVersionName = "Go"
	appVersionCode = "200117"

	copyrightYear = "2020"
	copyrightAuthor = "Aztec Rabbit"
)

func main() {
	SetupCloseHandler()

	liblog.LogColor(
		fmt.Sprintf("%s [%s Version. %s] \n(c) %s %s. \n",
			appName, appVersionName, appVersionCode, copyrightYear, copyrightAuthor),
		liblog.Colors["G1"],
	)

	var wg sync.WaitGroup

	threads := 6
	channel := make(chan bool, threads)

	Inject := new(libinject.Inject)
	Inject.Port = "8089"
	Inject.ProxyHost = "202.152.240.50"
	Inject.ProxyPort = "80"
	Inject.ProxyPayload = "CONNECT 103.253.27.56:80 HTTP/1.0\r\nHost: t.co\r\nHost: \r\n\r\n"
	Inject.ProxyTimeout = 5
	Inject.LogConnecting = false

	ProxyRotator := new(libproxyrotator.ProxyRotator)
	ProxyRotator.Port = "3080"

	if len(os.Args) > 1 {
		ProxyRotator.Port = os.Args[1]
	}

	for i := 1; i <= threads; i++ {
		wg.Add(1)

		ProxyRotatorPort, err := strconv.Atoi(ProxyRotator.Port)
		if err != nil {
			panic(err)
		}

		SshClient := new(sshclient.SshClient)
		// SshClient.Host = "m.sg1.ssh.speedssh.com"
		// SshClient.Host = "157.245.62.248"
		// SshClient.Port = "22"
		// SshClient.Username = "speedssh.com-aztecrabbit"
		SshClient.Host = "103.253.27.56"
		SshClient.Port = "80"
		SshClient.Username = "aztecrabbit"
		SshClient.Password = "aztecrabbit"
		SshClient.InjectHost = "0.0.0.0"
		SshClient.InjectPort = "8089"
		SshClient.ListenPort = strconv.Itoa(ProxyRotatorPort + i)
		SshClient.Loop = true

		ProxyRotator.Proxies = append(ProxyRotator.Proxies, "0.0.0.0:" + SshClient.ListenPort)

		go SshClient.Start(&wg, channel)
	}

	go ProxyRotator.Start()
	go Inject.Start()

	time.Sleep(200 * time.Millisecond)

	liblog.LogInfo("Proxy Rotator running on port " + ProxyRotator.Port, "INFO", liblog.Colors["G1"])
	liblog.LogInfo("Inject running on port " + Inject.Port, "INFO", liblog.Colors["G1"])

	for i := 0; i < threads; i++ {
		channel <- true
	}

	wg.Wait()
}

func SetupCloseHandler() {
    channelSignal := make(chan os.Signal, 2)
    signal.Notify(channelSignal, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- channelSignal
        liblog.LogInfo(
        	"Keyboard Interrupt\n\n" +
        		"|   Ctrl-C again if not exiting automaticly\n" +
        		"|   Please wait...\n|\n",

        	"INFO", liblog.Colors["R1"],
        )
        sshclient.Stop()
        os.Exit(0)
    }()
}
