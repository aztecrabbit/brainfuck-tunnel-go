package libsshclient

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
    ConfigDefault = &Config{
		Host: "157.245.62.248",
		Port: "22",
		Username: "speedssh.com-aztecrabbit",
		Password: "aztecrabbit",
		ProxyHost: "127.0.0.1",
		ProxyPort: "8989",
    }
)

func Stop() {
	Loop = false
}

type Config struct {
	Host string
	Port string
	Username string
	Password string
	ProxyHost string
	ProxyPort string
}

type SshClient struct {
	Config *Config
	ListenPort string
	Verbose bool
	Loop bool
}

func (s *SshClient) LogInfo(message string, color string) {
	if s.Loop {
		liblog.LogInfo(message, s.ListenPort, color)
	}
}

func (s *SshClient) Stop() {
	s.Loop = false
}

func (s *SshClient) Start(wg *sync.WaitGroup, channel chan bool) {
	defer wg.Done()

	<- channel

	s.LogInfo(fmt.Sprintf("Connecting to %s port %s", s.Config.Host, s.Config.Port), liblog.Colors["G1"])

	for Loop && s.Loop {
		command := exec.Command(
			"dash", "-c", fmt.Sprintf(
				"sshpass -p '%s' ssh -v %s -p %s -l '%s' " +
					"-o StrictHostKeyChecking=no " +
					"-o UserKnownHostsFile=/dev/null " +
					"-o ProxyCommand='corkscrew %s %s %%h %%p' " +
					// "-o ProxyCommand='nc -X CONNECT -x %s:%s %%h %%p' " +
					"-CND %s ",
				s.Config.Password,
				s.Config.Host,
				s.Config.Port,
				s.Config.Username,
				s.Config.ProxyHost,
				s.Config.ProxyPort,
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
					s.LogInfo("Connected", liblog.Colors["Y1"])

				} else if strings.Contains(line, "Permission denied") {
					s.LogInfo("Access Denied", liblog.Colors["R1"])
					s.Stop()

				} else if strings.Contains(line, "Connection closed") {
					s.LogInfo("Connection closed", liblog.Colors["R1"])

				} else if strings.Contains(line, "Address already in use") {
					s.LogInfo("Port used by another programs", liblog.Colors["R1"])
					s.Stop()

				} else {
					if s.Verbose {
						s.LogInfo(line, liblog.Colors["G2"])
					}
				}
			}
		}()

		command.Start()
		command.Wait()

		time.Sleep(200 * time.Millisecond)
	}

}
