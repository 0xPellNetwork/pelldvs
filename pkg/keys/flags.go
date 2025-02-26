package keys

type StringFlag struct {
	Name     string
	Aliases  string
	Value    string
	Usage    string
	EnvVars  []string
	Required bool
}

type BoolFlag struct {
	Name     string
	Aliases  string
	Value    bool
	Usage    string
	EnvVars  []string
	Required bool
}

var (
	KeyTypeFlag = StringFlag{
		Name:     "key-type",
		Aliases:  "k",
		Required: true,
		Usage:    "Type of key you want to create. Currently supports 'ecdsa' and 'bls'",
		EnvVars:  []string{"KEY_TYPE"},
	}

	InsecureFlag = BoolFlag{
		Name:    "insecure",
		Aliases: "i",
		Usage:   "Use this flag to skip password validation",
		EnvVars: []string{"INSECURE"},
	}

	KeyPathFlag = StringFlag{
		Name:    "key-path",
		Aliases: "p",
		Usage:   "Use this flag to specify the path of the key",
		EnvVars: []string{"KEY_PATH"},
	}
)
