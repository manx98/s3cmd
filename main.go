package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	appID   = 2040
	appHash = "b18441a1ff607e10a989891a5462e627"
)

var client *minio.Client
var configPath = flag.String("c", "config.ini", "config")
var commands = make(map[string]*Command)

const (
	helpCmd = "h"
)

type TerminalVars struct {
	Ctx context.Context
}

func (t *TerminalVars) Confirm(tip string) bool {
	for {
		fmt.Print(tip + " (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, " \r\n")
		switch text {
		case "y", "Y":
			return true
		case "n", "N":
			return false
		default:
			fmt.Println("请输入y或n")
		}
	}
}

type CommandHandler func(vars *TerminalVars, param string) bool

type Command struct {
	Desc    string
	Handler CommandHandler
}

func getTime() string {
	currentTime := time.Now()

	// 格式化为 "2006-01-02 15:04:05" 样式
	return currentTime.Format("2006-01-02 15:04:05")
}

func RegisterCommand(cmd string, desc string, handler CommandHandler) {
	commands[cmd] = &Command{Desc: desc, Handler: handler}
}

func Help(_ *TerminalVars, _ string) bool {
	for cmd, obj := range commands {
		fmt.Println(cmd + "\t" + obj.Desc)
	}
	return true
}

func ClearScreen(_ *TerminalVars, _ string) bool {
	var err1 error
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		err1 = cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err1 = cmd.Run()
	}
	if err1 != nil {
		fmt.Println(err1)
	}
	return true
}

func CmdLoop() {
	loop := true
	reader := bufio.NewReader(os.Stdin)
	vars := &TerminalVars{}
	vars.Ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGTERM)
	var err error
	for loop && vars.Ctx.Err() == nil {
		fmt.Print(getTime() + " > ")
		var cmd string
		cmd, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		cmd = strings.Trim(cmd, " \r\n")
		var params string
		cmdList := strings.SplitN(cmd, " ", 2)
		if len(cmdList) > 1 {
			params = strings.Trim(cmdList[1], " ")
		}
		if len(cmdList) > 0 {
			cmd = strings.Trim(cmdList[0], " ")
		} else {
			cmd = ""
		}
		if cmd != "" {
			if command, ok := commands[cmd]; ok {
				loop = command.Handler(vars, params)
			} else {
				commands[helpCmd].Handler(vars, "")
			}
		}
	}
}

func main() {
	flag.Parse()
	if *configPath == "" {
		flag.Usage()
		return
	}
	var err error
	var clientConfig *MinioConfig
	clientConfig, err = loadConfig()
	if err != nil {
		log.Fatalln(err)
	}
	var transport http.RoundTripper
	if clientConfig.SkipSSL {
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableCompression: true,
		}
	}
	client, err = minio.New(clientConfig.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(clientConfig.AccessKey, clientConfig.SecretKey, ""),
		Secure:    clientConfig.UseSSL,
		Transport: transport,
	})
	if err != nil {
		log.Fatalln(err)
	}
	RegisterCommand("clear", "清屏", ClearScreen)
	RegisterCommand(helpCmd, "帮助", Help)
	RegisterCommand("q", "退出 Ctrl+C", func(_ *TerminalVars, _ string) bool {
		return false
	})
	CmdLoop()
}
