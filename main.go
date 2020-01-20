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
	"github.com/aztecrabbit/brainfuck-tunnel-go/src/sshclient"
)

const (
	appName = "Brainfuck Tunnel"
	appVersionName = "Go"
	appVersionCode = "200120"

	copyrightYear = "2020"
	copyrightAuthor = "Aztec Rabbit"
)

func main() {
	InterruptHandler := &libutils.InterruptHandler{
		Handle: func() {
			sshclient.Stop()
			liblog.LogKeyboardInterrupt()
		},
    }
	InterruptHandler.Start()

	liblog.Header(
		[]string{
			fmt.Sprintf("%s [%s Version. %s]", appName, appVersionName, appVersionCode),
			fmt.Sprintf("(c) %s %s.", copyrightYear, copyrightAuthor),
		},
		liblog.Colors["G1"],
	)

	var wg sync.WaitGroup

	threads := 4
	channel := make(chan bool, threads)

	Inject := new(libinject.Inject)
	Inject.Port = "8089"
	Inject.ProxyHost = "202.152.240.50"
	Inject.ProxyPort = "80"
	Inject.ProxyPayload = "CONNECT [host_port] [protocol][crlf]Host: t.co[crlf]Host: [crlf][crlf]"
	Inject.ProxyTimeout = 10
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
