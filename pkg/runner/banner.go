package runner

import "github.com/projectdiscovery/gologger"

const (
	banner = `
                             ______      ____        _     __     
  ____ __________ ___  ___  / ____/___  / __ )____ _(_)___/ /_  __
 / __  / ___/ __  __ \/ _ \/ / __/ __ \/ __  / __  / / __  / / / /
/ /_/ / /__/ / / / / /  __/ /_/ / /_/ / /_/ / /_/ / / /_/ / /_/ /
\__,_/\___/_/ /_/ /_/\___/\____/\____/_____/\__,_/_/\__,_/\__,_/
`
	Version  = `1.0.0`
	userName = "wjlin0"
	repoName = "acmeGoBaidu"
)

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\t\twjlin0.com\n\n")
	gologger.Print().Msgf("慎用。你要为自己的行为负责\n")
	gologger.Print().Msgf("开发者不承担任何责任，也不对任何误用或损坏负责.\n")

}
