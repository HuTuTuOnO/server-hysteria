package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/xflash-panda/server-hysteria/internal/app"
	"github.com/xflash-panda/server-hysteria/internal/app/service"
	"github.com/xflash-panda/server-hysteria/internal/pkg/api"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	Name          = "hysteria-node"
	Version       = "0.1.2"
	CopyRight     = "XFLASH-PANDA@2021"
	LogLevelDebug = "debug"
	LogLevelError = "error"
	LogLevelInfo  = "info"
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}
	cli.ErrWriter = ioutil.Discard

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version=%s xray.version=%s\n", Version)
	}
}

func main() {
	var serverConfig app.ServerConfig
	var apiConfig api.Config
	var serviceConfig service.Config
	var logLevel string

	app := &cli.App{
		Name:      Name,
		Version:   Version,
		Copyright: CopyRight,
		Usage:     "Provide hysteria service for the v2Board(XFLASH-PANDA)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "api",
				Usage:       "Server address",
				EnvVars:     []string{"X_PANDA_HYSTERIA_API", "API"},
				Required:    true,
				Destination: &apiConfig.APIHost,
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "Token of server API",
				EnvVars:     []string{"X_PANDA_HYSTERIA_TOKEN", "TOKEN"},
				Required:    true,
				Destination: &apiConfig.Token,
			},
			&cli.IntFlag{
				Name:        "node",
				Usage:       "Node ID",
				EnvVars:     []string{"X_PANDA_HYSTERIA_NODE", "NODE"},
				Required:    true,
				Destination: &apiConfig.NodeID,
			},
			&cli.StringFlag{
				Name:        "cert_file",
				Usage:       "Cert file",
				EnvVars:     []string{"X_PANDA_HYSTERIA_CERT_FILE", "CERT_FILE"},
				Value:       "/root/.cert/server.crt",
				Required:    false,
				DefaultText: "/root/.cert/server.crt",
				Destination: &serverConfig.CertFile,
			},
			&cli.StringFlag{
				Name:        "key_file",
				Usage:       "Key file",
				EnvVars:     []string{"X_PANDA_HYSTERIA_KEY_FILE", "KEY_FILE"},
				Value:       "/root/.cert/server.key",
				Required:    false,
				DefaultText: "/root/.cert/server.key",
				Destination: &serverConfig.KeyFile,
			},
			&cli.DurationFlag{
				Name:        "fetch_users_interval",
				Usage:       "API request cycle(fetch users), unit: second",
				EnvVars:     []string{"X_PANDA_HYSTERIA_FETCH_USERS_INTERVAL", "FETCH_USERS_INTERVAL"},
				Value:       time.Second * 60,
				DefaultText: "60",
				Required:    false,
				Destination: &serviceConfig.FetchUserInterval,
			},
			&cli.DurationFlag{
				Name:        "report_traffics_interval",
				Usage:       "API request cycle(report traffics), unit: second",
				EnvVars:     []string{"X_PANDA_HYSTERIA_REPORT_TRAFFIC_INTERVAL", "REPORT_TRAFFICS_INTERVAL"},
				Value:       time.Second * 90,
				DefaultText: "60",
				Required:    false,
				Destination: &serviceConfig.ReportTrafficInterval,
			},
			&cli.StringFlag{
				Name:        "log_mode",
				Value:       LogLevelError,
				Usage:       "Log mode",
				EnvVars:     []string{"X_PANDA_HYSTERIA_LOG_MODE", "LOG_MODE"},
				Destination: &logLevel,
				Required:    false,
			},
		},
		Before: func(c *cli.Context) error {
			log.SetFormatter(&log.TextFormatter{})
			if logLevel == LogLevelDebug {
				log.SetFormatter(&log.TextFormatter{
					FullTimestamp: true,
				})
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
			} else if logLevel == LogLevelInfo {
				log.SetLevel(log.InfoLevel)
			} else if logLevel == LogLevelError {
				log.SetLevel(log.ErrorLevel)
			} else {
				return fmt.Errorf("log mode %s not supported", logLevel)
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			if logLevel != LogLevelDebug {
				defer func() {
					if e := recover(); e != nil {
						panic(e)
					}
				}()
			}
			apiClient := api.New(&apiConfig)
			nodeInfo, err := apiClient.GetNodeInfo()
			if err != nil {
				log.Fatalf("get node config error:%s", err)
			}
			serverConfig.DisableMTUDiscovery = nodeInfo.DisableMTUDiscovery != 0
			serverConfig.Protocol = nodeInfo.Protocol
			serverConfig.Obfs = nodeInfo.Obfs
			//与客户端取反
			serverConfig.UpMbps = nodeInfo.DownMbps
			serverConfig.DownMbps = nodeInfo.UpMbps
			serverConfig.Listen = fmt.Sprintf(":%d", nodeInfo.ServerPort)

			if err := serverConfig.Check(); err != nil {
				log.Fatalf("server config error: %s", err)
			}

			usersService := service.NewUsersService(&serviceConfig, apiClient)
			go app.Run(&serverConfig, usersService)
			osSignals := make(chan os.Signal, 1)
			signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
			for {
				runtime.GC()
				<-osSignals
				log.Infoln("server will close..")
				break
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
