package main

import (
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/alecthomas/kong"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/Sora233/DDBOT/logging"
	_ "github.com/Sora233/DDBOT/lsp"
)

func init() {
	utils.WriteLogToFS()
}

func main() {
	var cli struct {
		Play           bool  `optional:"" help:"run the play function"`
		Debug          bool  `optional:"" help:"enable debug mode"`
		GenerateDevice bool  `optional:"" xor:"c" help:"generate device.json"`
		SetAdmin       int64 `optional:"" xor:"c" help:"set QQ number to Admin"`
		Version        bool  `optional:"" xor:"c" short:"v" help:"print the version info"`
	}
	kong.Parse(&cli)

	if cli.Version {
		fmt.Printf("COMMIT_ID: %v\n", CommitId)
		fmt.Printf("BUILD_TIME: %v\n", BuildTime)
		os.Exit(0)
	}

	if cli.GenerateDevice {
		bot.GenRandomDevice()
		os.Exit(0)
	}

	if cli.SetAdmin != 0 {
		if err := localdb.InitBuntDB(""); err != nil {
			fmt.Println("can not init buntdb")
			os.Exit(1)
		}
		defer localdb.Close()
		sm := permission.NewStateManager()
		err := sm.GrantRole(cli.SetAdmin, permission.Admin)
		if err != nil {
			fmt.Printf("set role failed %v\n", err)
			os.Exit(1)
		}
		return
	}

	config.Init()

	// 快速初始化
	bot.Init()

	if cli.Debug {
		lsp.Debug = true
		go http.ListenAndServe("localhost:6060", nil)
	}

	if cli.Play {
		play()
		return
	}

	// 初始化 Modules
	bot.StartService()

	// 使用协议
	// 不同协议可能会有部分功能无法使用
	// 在登陆前切换协议
	bot.UseProtocol(bot.AndroidPhone)

	// 登录
	bot.Login()

	// 刷新好友列表，群列表
	bot.RefreshList()

	lsp.Instance.PostStart(bot.Instance)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ch
	bot.Stop()
}
