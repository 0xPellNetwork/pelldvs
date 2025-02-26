package config

var (
	CmtConfig *Config
)

func SetGlobalCmtConfig(cmtConfig *Config) {
	CmtConfig = cmtConfig
}
