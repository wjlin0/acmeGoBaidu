package types

type Options struct {
	ConfigFile         string `json:"configFile,omitempty"`
	JsonPath           string `json:"jsonPath,omitempty"`
	Cron               string `json:"cron,omitempty"`
	Version            bool   `json:"version,omitempty"`
	DisableUpdateCheck bool   `json:"disableUpdateCheck,omitempty"`
}
