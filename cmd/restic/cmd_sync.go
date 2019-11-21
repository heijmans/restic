package main

import (
	"fmt"
	"github.com/restic/restic/internal/ui/termstatus"
	"github.com/spf13/cobra"
	tomb "gopkg.in/tomb.v2"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

var cmdUp = &cobra.Command{
	Use:   "up [dir]",
	Short: "backup based on .restic.yaml config",
	Long: `
backup based on .restic.yaml config.
`,
	DisableAutoGenTag: true,
	RunE:              runUp,
}

type upOptions struct {
}

var upOpts upOptions

var cmdDown = &cobra.Command{
	Use:   "down [dir]",
	Short: "restore based on .restic.yaml config",
	Long: `
restore based on .restic.yaml config.
`,
	DisableAutoGenTag: true,
	RunE:              runDown,
}

type downOptions struct {
}

var downOpts downOptions

func init() {
	cmdRoot.AddCommand(cmdUp)
	cmdRoot.AddCommand(cmdDown)
}

type ResticConfig struct {
	Host string `yaml:"host"`
	Remote string `yaml:"remote"`
	Excludes []string `yaml:"excludes"`
}

func readConfig(dir string) (*ResticConfig, error) {
	path := filepath.Join(dir, ".restic.yaml")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while reading config %v: %v", path, err)
	}

	var config ResticConfig
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("error while parsing config %v: %v", path, err)
	}
	return &config, nil
}

func runUp(cmd *cobra.Command, args []string) error {
	n := len(args)
	var dir string
	if n == 0 {
		dir, _ = os.Getwd()
	} else if n == 1 {
		dir = args[0]
	} else {
		return fmt.Errorf("need 0 or 1 args, not %v", n)
	}

	config, err := readConfig(dir)
	if err != nil {
		return fmt.Errorf("error while reading config: %v", err)
	}

	if config.Remote != "" {
		globalOptions.Repo = config.Remote
	}
	if config.Host != "" {
		backupOptions.Host = config.Host
	}
	if config.Excludes != nil {
		backupOptions.Excludes = config.Excludes
	}
	backupOptions.IgnoreInode = true
	backupOptions.Root = dir
	backupArgs := []string{ "/" }

	var t tomb.Tomb
	term := termstatus.New(globalOptions.stdout, globalOptions.stderr, globalOptions.Quiet)
	t.Go(func() error { term.Run(t.Context(globalOptions.ctx)); return nil })

	err = runBackup(backupOptions, globalOptions, term, backupArgs)
	if err != nil {
		return err
	}
	t.Kill(nil)
	return t.Wait()
}

func runDown(cmd *cobra.Command, args []string) error {
	n := len(args)
	var dir string
	if n == 0 {
		dir, _ = os.Getwd()
	} else if n == 1 {
		dir = args[0]
	} else {
		return fmt.Errorf("need 0 or 1 args, not %v", n)
	}

	config, err := readConfig(dir)
	if err != nil {
		return fmt.Errorf("error while reading config: %v", err)
	}

	if config.Remote != "" {
		globalOptions.Repo = config.Remote
	}
	if config.Host != "" {
		restoreOptions.Host = config.Host
	}
	restoreOptions.Target = dir
	restoreArgs := []string{"latest"}

	return runRestore(restoreOptions, globalOptions, restoreArgs)
}
