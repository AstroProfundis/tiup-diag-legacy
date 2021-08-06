// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/joomcode/errorx"
	"github.com/pingcap/tidb-foresight/collector"
	"github.com/pingcap/tidb-foresight/version"
	"github.com/pingcap/tiup/pkg/cluster/executor"
	operator "github.com/pingcap/tiup/pkg/cluster/operation"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	tiupmeta "github.com/pingcap/tiup/pkg/environment"
	"github.com/pingcap/tiup/pkg/localdata"
	"github.com/pingcap/tiup/pkg/logger"
	"github.com/pingcap/tiup/pkg/repository"
	"github.com/pingcap/tiup/pkg/tui"
	"github.com/pingcap/tiup/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	rootCmd     *cobra.Command
	gOpt        operator.Options
	skipConfirm bool
)

var tidbSpec *spec.SpecManager
var cm *collector.Manager

func init() {
	logger.InitGlobalLogger()

	tui.AddColorFunctionsForCobra()

	cobra.EnableCommandSorting = false

	nativeEnvVar := strings.ToLower(os.Getenv(localdata.EnvNameNativeSSHClient))
	if nativeEnvVar == "true" || nativeEnvVar == "1" || nativeEnvVar == "enable" {
		gOpt.NativeSSH = true
	}

	rootCmd = &cobra.Command{
		Use:           tui.OsArgs0(),
		Short:         "Collect metrics and information from a TiDB cluster",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var env *tiupmeta.Environment
			// unset component data dir to use clusters'
			os.Unsetenv(localdata.EnvNameComponentDataDir)
			if err = spec.Initialize("cluster"); err != nil {
				return err
			}

			tidbSpec = spec.GetSpecManager()
			cm = collector.NewManager("tidb", tidbSpec, spec.TiDBComponentVersion)

			// Running in other OS/ARCH Should be fine we only download manifest file.
			env, err = tiupmeta.InitEnv(repository.Options{
				GOOS:   "linux",
				GOARCH: "amd64",
			})
			if err != nil {
				return err
			}
			tiupmeta.SetGlobalEnv(env)

			if gOpt.NativeSSH {
				gOpt.SSHType = executor.SSHTypeSystem
				zap.L().Info("System ssh client will be used",
					zap.String(localdata.EnvNameNativeSSHClient, os.Getenv(localdata.EnvNameNativeSSHClient)))
				fmt.Println("The --native-ssh flag has been deprecated, please use --ssh=system")
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return tiupmeta.GlobalEnv().V1Repository().Mirror().Close()
		},
	}

	tui.BeautifyCobraUsageAndHelp(rootCmd)

	rootCmd.PersistentFlags().Uint64Var(&gOpt.SSHTimeout, "ssh-timeout", 5, "Timeout in seconds to connect host via SSH, ignored for operations that don't need an SSH connection.")
	// the value of wait-timeout is also used for `systemctl` commands, as the default timeout of systemd for
	// start/stop operations is 90s, the default value of this argument is better be longer than that
	rootCmd.PersistentFlags().Uint64Var(&gOpt.OptTimeout, "wait-timeout", 180, "Timeout in seconds to wait for an operation to complete, ignored for operations that don't fit.")
	rootCmd.PersistentFlags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip all confirmations and assumes 'yes'")
	rootCmd.PersistentFlags().BoolVar(&gOpt.NativeSSH, "native-ssh", gOpt.NativeSSH, "(EXPERIMENTAL) Use the native SSH client installed on local system instead of the build-in one.")
	rootCmd.PersistentFlags().StringVar((*string)(&gOpt.SSHType), "ssh", "", "(EXPERIMENTAL) The executor type: 'builtin', 'system', 'none'.")
	rootCmd.PersistentFlags().IntVarP(&gOpt.Concurrency, "concurrency", "c", 5, "max number of parallel tasks allowed")
	_ = rootCmd.PersistentFlags().MarkHidden("native-ssh")

	rootCmd.AddCommand(
		newCollectCmd(),
		newRebuildCmd(),
	)
}

func printErrorMessageForNormalError(err error) {
	_, _ = tui.ColorErrorMsg.Fprintf(os.Stderr, "\nError: %s\n", err.Error())
}

func printErrorMessageForErrorX(err *errorx.Error) {
	msg := ""
	ident := 0
	causeErrX := err
	for causeErrX != nil {
		if ident > 0 {
			msg += strings.Repeat("  ", ident) + "caused by: "
		}
		currentErrMsg := causeErrX.Message()
		if len(currentErrMsg) > 0 {
			if ident == 0 {
				// Print error code only for top level error
				msg += fmt.Sprintf("%s (%s)\n", currentErrMsg, causeErrX.Type().FullName())
			} else {
				msg += fmt.Sprintf("%s\n", currentErrMsg)
			}
			ident++
		}
		cause := causeErrX.Cause()
		if c := errorx.Cast(cause); c != nil {
			causeErrX = c
		} else {
			if cause != nil {
				if ident > 0 {
					// The error may have empty message. In this case we treat it as a transparent error.
					// Thus `ident == 0` can be possible.
					msg += strings.Repeat("  ", ident) + "caused by: "
				}
				msg += fmt.Sprintf("%s\n", cause.Error())
			}
			break
		}
	}
	_, _ = tui.ColorErrorMsg.Fprintf(os.Stderr, "\nError: %s", msg)
}

func extractSuggestionFromErrorX(err *errorx.Error) string {
	cause := err
	for cause != nil {
		v, ok := cause.Property(utils.ErrPropSuggestion)
		if ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		cause = errorx.Cast(cause.Cause())
	}

	return ""
}

// Execute executes the root command
func Execute() {
	zap.L().Info("Execute command", zap.String("command", tui.OsArgs()))
	zap.L().Debug("Environment variables", zap.Strings("env", os.Environ()))

	// Switch current work directory if running in TiUP component mode
	if wd := os.Getenv(localdata.EnvNameWorkDir); wd != "" {
		if err := os.Chdir(wd); err != nil {
			zap.L().Warn("Failed to switch work directory", zap.String("working_dir", wd), zap.Error(err))
		}
	}

	code := 0
	err := rootCmd.Execute()
	if err != nil {
		code = 1
	}

	zap.L().Info("Execute command finished", zap.Int("code", code), zap.Error(err))

	if err != nil {
		if errx := errorx.Cast(err); errx != nil {
			printErrorMessageForErrorX(errx)
		} else {
			printErrorMessageForNormalError(err)
		}

		if !errorx.HasTrait(err, utils.ErrTraitPreCheck) {
			logger.OutputDebugLog("tiup-cluster")
		}

		if errx := errorx.Cast(err); errx != nil {
			if suggestion := extractSuggestionFromErrorX(errx); len(suggestion) > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "\n%s\n", suggestion)
			}
		}
	}

	err = logger.OutputAuditLogIfEnabled()
	if err != nil {
		zap.L().Warn("Write audit log file failed", zap.Error(err))
		code = 1
	}

	color.Unset()

	if code != 0 {
		os.Exit(code)
	}
}