package sshclient

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
	"bufio"
	"strings"
	"github.com/aztecrabbit/liblog"
)

var (
	Loop bool = true
)

func Stop() {
	Loop = false
}

type SshClient struct {
	Host string
	Port string
	Username string
	Password string
	InjectHost string
	InjectPort string
	ListenPort string
	Loop bool
}

func (s *SshClient) LogColor(message string, color string) {
	if s.Loop {
		liblog.LogInfo(message, s.ListenPort, color)
	}
}

func (s *SshClient) Log(message string) {
	s.LogColor(message, liblog.Colors["G1"])
}

func (s *SshClient) Stop() {
	s.Loop = false
}

func (s *SshClient) Start(wg *sync.WaitGroup, channel chan bool) {
	defer wg.Done()

	<- channel

	s.Log(fmt.Sprintf("Connecting to %s port %s", s.Host, s.Port))

	for Loop && s.Loop {
		command := exec.Command(
			"dash", "-c", fmt.Sprintf(
				"sshpass -p '%s' ssh -v %s -p %s -l '%s' " +
					"-o StrictHostKeyChecking=no " +
					"-o UserKnownHostsFile=/dev/null " +
					"-o ProxyCommand='corkscrew %s %s %%h %%p' " +
					// "-o ProxyCommand='nc -X CONNECT -x %s:%s %%h %%p' " +
					"-CND %s ",
				s.Password,
				s.Host,
				s.Port,
				s.Username,
				s.InjectHost,
				s.InjectPort,
				s.ListenPort,
			),
		)

		stderr, err := command.StderrPipe()
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		go func() {
			var line string
			for Loop && s.Loop && scanner.Scan() {
				line = scanner.Text()

				if strings.Contains(line, "debug1: pledge: ") {
					s.LogColor("Connected", liblog.Colors["Y1"])

				} else if strings.Contains(line, "Permission denied") {
					s.LogColor("Access Denied", liblog.Colors["R1"])
					s.Stop()

				} else if strings.Contains(line, "Connection closed") {
					s.LogColor("Connection closed", liblog.Colors["R1"])

				} else if strings.Contains(line, "Address already in use") {
					s.LogColor("Port used by another programs", liblog.Colors["R1"])
					s.Stop()

				} else {
					// s.LogColor(line, liblog.Colors["G2"])

				}
			}
		}()

		command.Start()
		command.Wait()

		time.Sleep(200 * time.Millisecond)
	}

}
