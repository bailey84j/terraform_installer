package main

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/bailey84j/terraform_installer/pkg/asset"
	"github.com/bailey84j/terraform_installer/pkg/asset/cluster"

	assetstore "github.com/bailey84j/terraform_installer/pkg/asset/store"

	"github.com/bailey84j/terraform_installer/pkg/asset/logging"
	timer "github.com/bailey84j/terraform_installer/pkg/metrics/timer"
)

type target struct {
	name    string
	command *cobra.Command
	assets  []asset.WritableAsset
}

const (
	exitCodeInstallConfigError = iota + 3
	exitCodeInfrastructureFailed
	exitCodeBootstrapFailed
	exitCodeInstallFailed
)

var (
	targets = []target{}
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create part of an Terraform Enterprise cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	for _, t := range targets {
		t.command.Args = cobra.ExactArgs(0)
		t.command.Run = runTargetCmd(t.assets...)
		cmd.AddCommand(t.command)
	}

	return cmd
}

func runTargetCmd(targets ...asset.WritableAsset) func(cmd *cobra.Command, args []string) {
	runner := func(directory string) error {
		assetStore, err := assetstore.NewStore(directory)
		if err != nil {
			return errors.Wrap(err, "failed to create asset store")
		}

		for _, a := range targets {
			err := assetStore.Fetch(a, targets...)
			if err != nil {
				err = errors.Wrapf(err, "failed to fetch %s", a.Name())
			}

			err2 := asFileWriter(a).PersistToFile(directory)
			if err2 != nil {
				err2 = errors.Wrapf(err2, "failed to write asset (%s) to disk", a.Name())
				if err != nil {
					logrus.Error(err2)
					return err
				}
				return err2
			}

			if err != nil {
				return err
			}
		}
		return nil
	}

	return func(cmd *cobra.Command, args []string) {
		timer.StartTimer(timer.TotalTimeElapsed)

		cleanup := setupFileHook(rootOpts.dir)
		defer cleanup()

		cluster.InstallDir = rootOpts.dir

		err := runner(rootOpts.dir)
		if err != nil {
			if strings.Contains(err.Error(), asset.InstallConfigError) {
				logrus.Error(err)
				logrus.Exit(exitCodeInstallConfigError)
			}
			if strings.Contains(err.Error(), asset.ClusterCreationError) {
				logrus.Error(err)
				logrus.Exit(exitCodeInfrastructureFailed)
			}
			logrus.Fatal(err)
		}
		switch cmd.Name() {
		case "cluster", "image":
		default:
			logrus.Infof(logging.LogCreatedFiles(cmd.Name(), rootOpts.dir, targets))
		}

	}
}

func asFileWriter(a asset.WritableAsset) asset.FileWriter {
	switch v := a.(type) {
	case asset.FileWriter:
		return v
	default:
		return asset.NewDefaultFileWriter(a)
	}
}
