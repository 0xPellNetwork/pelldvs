package config

var (
	GlobalPellConfig *PellConfig
)

func SetGlobalPellConfig(pellConfig *PellConfig) {
	GlobalPellConfig = pellConfig
}
