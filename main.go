package main

import (
	"os"
	"os/signal"
	"fmt"
	"sync"
	"syscall"
	"strconv"

	"github.com/aztecrabbit/liblog"
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
		SshClient.InjectPort = "8989"
		SshClient.ListenPort = strconv.Itoa(ProxyRotatorPort + i)
		SshClient.Loop = true

		ProxyRotator.Proxies = append(ProxyRotator.Proxies, "0.0.0.0:" + SshClient.ListenPort)

		go SshClient.Start(&wg, channel)
	}

	ProxyRotator.Start()

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
