package app

import (
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func SetConfig() ([]*core.Payload, bool, error) {
	var (
		configPath   string
		runMode      = internal.CliMode
		policy       = internal.PlcNotify
		env          = "development"
		port         = 6520
		head         string
		srcs         string
		descriptions string
	)

	var payloads []*core.Payload

	rootCmd := &cobra.Command{
		Use:           "clonopher",
		Short:         "Organize files with Clonopher.",
		Long:          `Clonopher organizes files using configurable classifiers and destination policies.`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			// Default configuration.
			// config.Default.App.RunMode = internal.ApiMode
			// config.Default.App.ActionPolicy = internal.PlcNotify
			// config.Default.App.Env = "development"
			// config.Default.App.Head = ""
			// config.Default.Api.Port = 6520

			if configPath != "" {
				config.Default.Setup(configPath)
			}

			flags := cmd.Flags()

			if flags.Changed("mode") {
				config.Default.App.RunMode = runMode
			}
			if flags.Changed("policy") {
				config.Default.App.ActionPolicy = policy
			}
			if flags.Changed("env") {
				config.Default.App.Env = env
			}
			if flags.Changed("port") {
				config.Default.Api.Port = port
			}
			// if flags.Changed("head") {
			// 	config.Default.App.Head = head
			// }

			payloads = SetPayload(srcs, descriptions, head)
			if len(payloads) > 0 {
				config.Default.App.RunMode = internal.CliMode
			}
			return nil
		},
	}

	flags := rootCmd.Flags()

	// General flags
	flags.StringVarP(
		&configPath, "config", "c", "",
		"Path to the configuration file.",
	)
	flags.StringVarP(
		&runMode, "mode", "m", internal.ApiMode,
		"Execution mode: cli or api.",
	)
	flags.StringVarP(
		&policy, "policy", "p", internal.PlcNotify,
		"Action policy: copy, move, notify, ask-copy, or ask-move.",
	)
	flags.StringVarP(
		&env, "env", "e", "development",
		"Application environment.",
	)

	// API flags
	flags.IntVarP(
		&port, "port", "P", 6520,
		"Port used by the API server.",
	)

	// CLI flags
	flags.StringVarP(
		&head, "head", "H", "",
		"Root directory for file organization.",
	)
	flags.StringVarP(
		&srcs, "src", "s", "",
		"Comma-separated source file paths.",
	)
	flags.StringVarP(
		&descriptions, "info", "i", "",
		"Comma-separated descriptions corresponding to the source paths. Use '-' to skip a description.",
	)

	if err := rootCmd.Execute(); err != nil {
		return nil, false, err
	}

	if rootCmd.Flags().Changed("help") {
		return nil, false, nil
	}

	return payloads, true, nil
}

func SetPayload(srcs string, descriptions string, head string) []*core.Payload {
	if srcs == "" {
		return nil
	}
	var res []*core.Payload
	srclist := strings.Split(srcs, ",")
	descList := strings.Split(descriptions, ",")
	if len(srclist) != len(descList) {
		loggy.Fatal("src(s) and description(s) are not match!, one has more parameters.")
	}
	for i := range len(srclist) {
		name := filepath.Base(srclist[i])
		p, err := core.New(&config.Default, name,
			head,
			srclist[i], descList[i],
			config.Default.App.ActionPolicy,
			config.Default.App.RunMode)
		if err != nil {
			loggy.Fatal(err)
		}
		if p != nil {
			res = append(res, p)
		}
	}
	return res
}
