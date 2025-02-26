package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	e2e "github.com/0xPellNetwork/pelldvs/test/e2e/pkg"
	"github.com/0xPellNetwork/pelldvs/test/e2e/pkg/infra"
	"github.com/0xPellNetwork/pelldvs/test/e2e/pkg/infra/digitalocean"
	"github.com/0xPellNetwork/pelldvs/test/e2e/pkg/infra/docker"
)

const randomSeed = 2308084734268

var logger = log.NewLogger(os.Stdout)

func main() {
	NewCLI().Run()
}

// CLI is the Cobra-based command-line interface.
type CLI struct {
	root     *cobra.Command
	testnet  *e2e.Testnet
	preserve bool
	infp     infra.Provider
}

// NewCLI sets up the CLI.
func NewCLI() *CLI {
	cli := &CLI{}
	cli.root = &cobra.Command{
		Use:           "runner",
		Short:         "End-to-end test runner",
		SilenceUsage:  true,
		SilenceErrors: true, // we'll output them ourselves in Run()
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}
			m, err := e2e.LoadManifest(file)
			if err != nil {
				return err
			}

			inft, err := cmd.Flags().GetString("infrastructure-type")
			if err != nil {
				return err
			}

			var ifd e2e.InfrastructureData
			switch inft {
			case "docker":
				var err error
				ifd, err = e2e.NewDockerInfrastructureData(m)
				if err != nil {
					return err
				}
			case "digital-ocean":
				p, err := cmd.Flags().GetString("infrastructure-data")
				if err != nil {
					return err
				}
				if p == "" {
					return errors.New("'--infrastructure-data' must be set when using the 'digital-ocean' infrastructure-type")
				}
				ifd, err = e2e.InfrastructureDataFromFile(p)
				if err != nil {
					return fmt.Errorf("parsing infrastructure data: %s", err)
				}
			default:
				return fmt.Errorf("unknown infrastructure type '%s'", inft)
			}

			testnet, err := e2e.LoadTestnet(file, ifd)
			if err != nil {
				return fmt.Errorf("loading testnet: %s", err)
			}

			cli.testnet = testnet
			switch inft {
			case "docker":
				cli.infp = &docker.Provider{
					ProviderData: infra.ProviderData{
						Testnet:            testnet,
						InfrastructureData: ifd,
					},
				}
			case "digital-ocean":
				cli.infp = &digitalocean.Provider{
					ProviderData: infra.ProviderData{
						Testnet:            testnet,
						InfrastructureData: ifd,
					},
				}
			default:
				return fmt.Errorf("bad infrastructure type: %s", inft)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Cleanup(cli.testnet); err != nil {
				return err
			}
			if err := Setup(cli.testnet, cli.infp); err != nil {
				return err
			}

			//r := rand.New(rand.NewSource(randomSeed)) //nolint: gosec

			chLoadResult := make(chan error)
			_, loadCancel := context.WithCancel(context.Background())
			defer loadCancel()

			if err := Start(cmd.Context(), cli.testnet, cli.infp); err != nil {
				return err
			}

			loadCancel()
			if err := <-chLoadResult; err != nil {
				return err
			}

			if err := Test(cli.testnet, cli.infp.GetInfrastructureData()); err != nil {
				return err
			}
			if !cli.preserve {
				if err := Cleanup(cli.testnet); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cli.root.PersistentFlags().StringP("file", "f", "", "Testnet TOML manifest")
	_ = cli.root.MarkPersistentFlagRequired("file")

	cli.root.PersistentFlags().StringP("infrastructure-type", "", "docker", "Backing infrastructure used to run the testnet. Either 'digital-ocean' or 'docker'")

	cli.root.PersistentFlags().StringP("infrastructure-data", "", "", "path to the json file containing the infrastructure data. Only used if the 'infrastructure-type' is set to a value other than 'docker'")

	cli.root.Flags().BoolVarP(&cli.preserve, "preserve", "p", false,
		"Preserves the running of the test net after tests are completed")

	cli.root.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "Generates the testnet directory and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Setup(cli.testnet, cli.infp)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts the testnet, waiting for nodes to become available",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := os.Stat(cli.testnet.Dir)
			if os.IsNotExist(err) {
				err = Setup(cli.testnet, cli.infp)
			}
			if err != nil {
				return err
			}
			return Start(cmd.Context(), cli.testnet, cli.infp)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stops the testnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Stopping testnet")
			return cli.infp.StopTestnet(context.Background())
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Runs test cases against a running testnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Test(cli.testnet, cli.infp.GetInfrastructureData())
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "cleanup",
		Short: "Removes the testnet directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Cleanup(cli.testnet)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "logs",
		Short: "Shows the testnet logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return docker.ExecComposeVerbose(context.Background(), cli.testnet.Dir, "logs")
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "tail",
		Short: "Tails the testnet logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return docker.ExecComposeVerbose(context.Background(), cli.testnet.Dir, "logs", "--follow")
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "benchmark",
		Short: "Benchmarks testnet",
		Long: `Benchmarks the following metrics:
	Mean Block Interval
	Standard Deviation
	Min Block Interval
	Max Block Interval
over a 100 block sampling period.

Does not run any perturbations.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Cleanup(cli.testnet); err != nil {
				return err
			}
			if err := Setup(cli.testnet, cli.infp); err != nil {
				return err
			}

			chLoadResult := make(chan error)

			if err := Start(cmd.Context(), cli.testnet, cli.infp); err != nil {
				return err
			}

			if err := <-chLoadResult; err != nil {
				return err
			}

			return Cleanup(cli.testnet)
		},
	})

	return cli
}

// Run runs the CLI.
func (cli *CLI) Run() {
	if err := cli.root.Execute(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
