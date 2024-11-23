package runner

import (
	"fmt"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/wjlin0/acmeGoBaidu/pkg/types"
	updateutils "github.com/wjlin0/utils/update"
	"os"
	"path"
)

func ParseOptions() *types.Options {
	options := &types.Options{}
	set := goflags.NewFlagSet()
	set.SetDescription(fmt.Sprintf("acmeGoBaidu %s Go编写的自动申请SSL证书并同步到百度CDN的工具", Version))
	set.CreateGroup("Input", "输入",
		set.StringVar(&options.ConfigFile, "config", "config/config.yaml", "配置文件"),
		set.StringVarP(&options.JsonPath, "json", "j", "certs/certificates.json", "证书信息存储文件"),
		set.StringVarP(&options.Cron, "cron", "c", "", "定时任务"),
	)
	set.CreateGroup("Version", "版本",
		set.BoolVarP(&options.Version, "version", "v", false, "显示版本信息"),
		set.CallbackVar(updateutils.GetUpdateToolCallback(repoName, Version), "update", "更新版本"),
		set.BoolVarP(&options.DisableUpdateCheck, "disable-update-check", "duc", false, "跳过自动检查更新"),
	)
	set.SetCustomHelpText(`EXAMPLES:

运行 acmeGoBaidu 单次运行:
    $ acmeGoBaidu

运行 acmeGoBaidu 并设置定时任务(e.g. 每天0点):
    $ nohup acmeGoBaidu -c "0 0 * * *" &
运行 acmeGoBaidu 使用环境变量指定配置文件:
    $ CONFIG_PATH=config.yaml JSON_PATH=certificates.json CRON="0 0 * * *" acmeGoBaidu
`)
	_ = set.Parse()

	// set default options
	EnvOptions(options)

	ValidateOptions(options)

	// show banner
	showBanner()
	// display version
	displayVersion(options)

	return options
}
func EnvOptions(options *types.Options) {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		options.ConfigFile = configPath
	}
	if jsonPath := os.Getenv("JSON_PATH"); jsonPath != "" {
		options.JsonPath = jsonPath
	}
	if cron := os.Getenv("CRON"); cron != "" {
		options.Cron = cron
	}
}

func ValidateOptions(options *types.Options) {
	if options.ConfigFile == "" {
		gologger.Fatal().Msgf("配置文件不能为空")
	}
	dir := path.Dir(options.ConfigFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			gologger.Fatal().Msgf("创建目录失败: %v", err)
		}
	}

	if options.JsonPath == "" {
		gologger.Fatal().Msgf("证书信息存储文件不能为空")
	}

	dir = path.Dir(options.JsonPath)

	// dir 是否存在 若不存在则创建
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			gologger.Fatal().Msgf("创建目录失败: %v", err)
		}
	}

}
func displayVersion(opts *types.Options) {
	if !opts.DisableUpdateCheck {
		latestVersion, err := updateutils.GetToolVersionCallback(repoName, repoName)()
		if err != nil {
			gologger.Error().Msgf("%s version check failed: %v", repoName, err.Error())

		} else {
			gologger.Info().Msgf("Current %s version v%v %v", repoName, Version, updateutils.GetVersionDescription(Version, latestVersion))
		}

	} else {
		gologger.Info().Msgf("Current %s version v%v ", repoName, Version)
	}
}
