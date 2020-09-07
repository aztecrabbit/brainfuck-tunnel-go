package libsshclient

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/aztecrabbit/liblog"
	"github.com/aztecrabbit/libproxyrotator"
)

var (
	Loop          = true
	DefaultConfig = &Config{
		Host:     "127.0.0.1",
		Port:     "22",
		Username: "root",
		Password: "toor",
	}
)

func Stop() {
	Loop = false
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

type SshClient struct {
	ProxyRotator *libproxyrotator.ProxyRotator
	Config       *Config
	InjectPort   string
	ListenPort   string
	Verbose      bool
	Loop         bool
}

func (s *SshClient) LogInfo(message string, color string) {
	if Loop && s.Loop {
		liblog.LogInfo(message, s.ListenPort, color)
	}
}

func (s *SshClient) Stop() {
	s.Loop = false
}

func (s *SshClient) Start() {
	s.LogInfo("Connecting", liblog.Colors["G1"])

	for Loop && s.Loop {
		command := exec.Command(
			"sh", "-c", fmt.Sprintf(
				"sshpass -p '%s' ssh -v %s -p %s -l '%s' "+
					"-o StrictHostKeyChecking=no "+
					"-o UserKnownHostsFile=/dev/null "+
					"-o ProxyCommand='corkscrew 127.0.0.1 %s %%h %%p' "+
					"-CND %s",
				s.Config.Password,
				s.Config.Host,
				s.Config.Port,
				s.Config.Username,
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

				if line == "debug1: Connection to port "+s.ListenPort+" forwarding to socks port 0 requested." {
					liblog.LogReplace(s.ListenPort, liblog.Colors["G1"])

				} else if strings.Contains(line, "debug1: pledge: ") {
					s.ProxyRotator.AddProxy("0.0.0.0:" + s.ListenPort)
					s.LogInfo("Connected", liblog.Colors["Y1"])

				} else if strings.Contains(line, "Permission denied") {
					s.LogInfo("Access Denied", liblog.Colors["R1"])
					s.Stop()

				} else if strings.Contains(line, "Connection closed by remote host") {
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

			command.Process.Signal(syscall.SIGTERM)
		}()

		command.Start()
		command.Wait()

		s.ProxyRotator.DeleteProxy("0.0.0.0:" + s.ListenPort)

		time.Sleep(200 * time.Millisecond)

		s.LogInfo("Reconnecting", liblog.Colors["G1"])
	}
}
