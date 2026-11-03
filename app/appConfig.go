package app

import (
	"clonopher/api"
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/fileSystem"
	"clonopher/internal/loggy"
	"context"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

var (
	loadedConfigByFlag  = false
	loadedDefaultConfig = false
	configPath          string
	roots               string
	srcs                string
	Destination         string
	runMode             = internal.CliMode
	policy              = internal.PlcNotify
	env                 = "development"
	port                = 6520
	infos               string
	logLevel            string
)

var rootCmd = &cobra.Command{
	Use:           "clonopher",
	Short:         "Organize files with Clonopher.",
	Long:          `Clonopher organizes files using configurable classifiers and destination policies.`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if configPath != "" {
			loadedConfigByFlag, err := config.Default.Setup(configPath)
			if !loadedConfigByFlag {
				if err != nil {
					return err
				}
				return loggy.Say("failed to load the flag-input config")
			}
		}

		if !loadedDefaultConfig && !loadedConfigByFlag {
			return loggy.Sayf("failed to Load a Config File, provide a default config here <%v> or use your own by -c or --config", filepath.Join(config.Default.App.AppConfigDir, internal.ConfigFileName))
		}

		flags := cmd.Flags()

		if flags.Changed("env") {
			config.Default.App.Env = env
		}
		if flags.Changed("roots") {
			err := ParseRoots(roots)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func setRootFlags() {
	var flags = rootCmd.PersistentFlags()

	// General flags
	flags.StringVarP(
		&configPath, "config", "c", "",
		"Path to the configuration file.",
	)
	flags.StringVarP(
		&roots, "roots", "r", "",
		"Comma-Separated Root directory for file organization. e.g: Root1Name:Root1Path,Root2Name:Root2Path,...",
	)
	flags.StringVarP(
		&env, "env", "e", "development",
		"Application environment.",
	)
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start Clonopher API server",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("port") {
			config.Default.Api.Port = port
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		apiConf := api.ApiConfig{
			Ctx:     ctx,
			Close:   cancel,
			Options: &config.Default,
		}
		apiConf.Run()

		return nil
	},
}

func setApiFlags() {
	// API flags
	flags := apiCmd.Flags()
	flags.IntVarP(
		&port, "port", "P", 6520,
		"Port used by the API server.",
	)
}

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Start Clonopher cli Organizer",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if flags.Changed("policy") {
			config.Default.App.ActionPolicy = policy
		}
		if flags.Changed("loglevel") {
			config.Default.Cli.LogLevel = logLevel
		}

		var finalDstRoot string

		if Destination == "" {
			if config.Default.App.DefaultRoot == "" {
				return loggy.Say(`no destination root was selected, you have to set a root by affing flag -d or --dst <a valid root name in config file or a path
				 or the root you want you file to be proccessed in or set a root-name as default by adding 'default_root' to app/ section inside your config.`)
			} else {
				ok, err := fileSystem.IsDir(config.Default.App.DefaultRoot)
				if ok {
					finalDstRoot = config.Default.App.DefaultRoot
				} else {
					if err != nil {
						return err
					}
					return loggy.Sayf("something went wrong while validating the destination Root: %v", config.Default.App.DefaultRoot)
				}
			}
		} else {
			var ok bool
			finalDstRoot, ok = config.Default.App.Roots[Destination]
			if !ok {
				finalDstRoot = Destination
			}
			ok, err := fileSystem.IsDir(finalDstRoot)
			if !ok {
				if err != nil {
					return err
				}
				return loggy.Sayf("something went wrong while validating the destination Root: %v", finalDstRoot)
			}
		}

		payloads, err := ParseSources(srcs, infos, finalDstRoot, config.Default.App.ActionPolicy)
		if err != nil {
			return err
		}

		done := make(chan struct{})
		tui, err := core.StartPayloadTUI(done, config.Default.Cli.LogLevel)
		if err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, p := range payloads {
			run := func() {
				_, err := p.Serve()
				if err != nil {
					loggy.Error(err)
				}
			}
			wg.Go(run)
		}

		wg.Wait()

		close(done)
		if err := tui.Wait(); err != nil {
			return err
		}

		return nil
	},
}

func setCliFlags() {
	// CLI flags
	flags := cliCmd.Flags()

	flags.StringVarP(
		&srcs, "src", "s", "",
		"Comma-separated source file paths. use () for multiple files in a row e.g: src1,(src2.1,src2.2,...),src3,...",
	)
	flags.StringVarP(
		&infos, "info", "i", "",
		"Comma-separated descriptions corresponding to the source paths. Use '-' to skip a description.",
	)
	flags.StringVarP(
		&Destination, "dst", "d", "",
		"destination Root for payload, should be a RootName from list or a path itself",
	)
	flags.StringVarP(
		&policy, "policy", "p", internal.PlcNotify,
		"Action policy: copy, move, notify, ask-copy, or ask-move.",
	)
	flags.StringVarP(
		&logLevel, "loglevel", "l", "",
		"loglevel for cli",
	)

}
