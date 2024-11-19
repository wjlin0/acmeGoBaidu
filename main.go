package main

import (
	"github.com/projectdiscovery/gologger"
	"github.com/robfig/cron/v3"
	"github.com/wjlin0/acmeGoBaidu/pkg/runner"
)

func main() {
	// 创建并运行 Runner
	runnerInstance, err := runner.NewRunner(runner.ParseOptions())
	if err != nil {
		gologger.Fatal().Msgf("初始化 Runner 失败: %v", err)
	}
	if runnerInstance.Options.Cron != "" {
		gologger.Info().Msgf("cron: %s", runnerInstance.Options.Cron)
		err = runnerInstance.Run()
		if err != nil {
			gologger.Fatal().Msgf("执行失败: %v", err)
		}

		c := cron.New()
		_, err = c.AddFunc(runnerInstance.Options.Cron, func() {
			err = runnerInstance.Run()
			if err != nil {
				gologger.Error().Msgf("执行失败: %v", err)
			}
		})
		if err != nil {
			gologger.Fatal().Msgf("添加任务失败: %v", err)
		}
		c.Start()
		// tail
		select {}
	} else {
		err = runnerInstance.Run()
		if err != nil {
			gologger.Fatal().Msgf("执行失败: %v", err)
		}
	}

}
