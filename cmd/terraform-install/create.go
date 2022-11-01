package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/bailey84j/terraform_installer/pkg/asset"
	"github.com/bailey84j/terraform_installer/pkg/asset/cluster"
	targetassets "github.com/bailey84j/terraform_installer/pkg/asset/targets"

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

// each target is a variable to preserve the order when creating subcommands and still
// allow other functions to directly access each target individually.
var (
	installConfigTarget = target{
		name: "Install Config",
		command: &cobra.Command{
			Use:   "install-config",
			Short: "Generates the Install Config asset",
			// FIXME: add longer descriptions for our commands with examples for better UX.
			// Long:  "",
		},
		assets: targetassets.InstallConfig,
	}

	manifestsTarget = target{
		name: "Manifests",
		command: &cobra.Command{
			Use:   "manifests",
			Short: "Generates the Kubernetes manifests",
			// FIXME: add longer descriptions for our commands with examples for better UX.
			// Long:  "",
		},
		assets: targetassets.Manifests,
	}

	ignitionConfigsTarget = target{
		name: "Ignition Configs",
		command: &cobra.Command{
			Use:   "ignition-configs",
			Short: "Generates the Ignition Config asset",
			// FIXME: add longer descriptions for our commands with examples for better UX.
			// Long:  "",
		},
		assets: targetassets.IgnitionConfigs,
	}
	singleNodeIgnitionConfigTarget = target{
		name: "Single Node Ignition Config",
		command: &cobra.Command{
			Use:   "single-node-ignition-config",
			Short: "Generates the bootstrap-in-place-for-live-iso Ignition Config asset",
			// FIXME: add longer descriptions for our commands with examples for better UX.
			// Long:  "",
		},
		assets: targetassets.SingleNodeIgnitionConfig,
	}

	clusterTarget = target{
		name: "Cluster",
		command: &cobra.Command{
			Use:   "cluster",
			Short: "Create an OpenShift cluster",
			// FIXME: add longer descriptions for our commands with examples for better UX.
			// Long:  "",
			PostRun: func(_ *cobra.Command, _ []string) {
				ctx := context.Background()

				cleanup := setupFileHook(rootOpts.dir)
				defer cleanup()

				// FIXME: pulling the kubeconfig and metadata out of the root
				// directory is a bit cludgy when we already have them in memory.
				config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(rootOpts.dir, "auth", "kubeconfig"))
				if err != nil {
					logrus.Fatal(errors.Wrap(err, "loading kubeconfig"))
				}

				timer.StartTimer("Bootstrap Complete")
				/*
					if err := waitForBootstrapComplete(ctx, config); err != nil {
						bundlePath, gatherErr := runGatherBootstrapCmd(rootOpts.dir)
						if gatherErr != nil {
							logrus.Error("Attempted to gather debug logs after installation failure: ", gatherErr)
						}
						//if err := logClusterOperatorConditions(ctx, config); err != nil {
						//	logrus.Error("Attempted to gather ClusterOperator status after installation failure: ", err)
						//}
						logrus.Error("Bootstrap failed to complete: ", err.Unwrap())
						logrus.Error(err.Error())
						if gatherErr == nil {
							//if err := service.AnalyzeGatherBundle(bundlePath); err != nil {
							logrus.Error("Attempted to analyze the debug logs after installation failure: ", err)
							//}
							logrus.Infof("Bootstrap gather logs captured here %q", bundlePath)
						}
						logrus.Exit(exitCodeBootstrapFailed)
					}*/
				timer.StopTimer("Bootstrap Complete")
				timer.StartTimer("Bootstrap Destroy")

				if oi, ok := os.LookupEnv("OPENSHIFT_INSTALL_PRESERVE_BOOTSTRAP"); ok && oi != "" {
					logrus.Warn("OPENSHIFT_INSTALL_PRESERVE_BOOTSTRAP is set, not destroying bootstrap resources. " +
						"Warning: this should only be used for debugging purposes, and poses a risk to cluster stability.")
				} else {
					logrus.Info("Destroying the bootstrap resources...")
					//err = destroybootstrap.Destroy(rootOpts.dir)
					if err != nil {
						logrus.Fatal(err)
					}
				}
				timer.StopTimer("Bootstrap Destroy")

				err = waitForInstallComplete(ctx, config, rootOpts.dir)
				if err != nil {
					//if err2 := logClusterOperatorConditions(ctx, config); err2 != nil {
					//	logrus.Error("Attempted to gather ClusterOperator status after installation failure: ", err2)
					//}
					//logTroubleshootingLink()
					logrus.Error(err)
					logrus.Exit(exitCodeInstallFailed)
				}
				timer.StopTimer(timer.TotalTimeElapsed)
				timer.LogSummary()
			},
		},
		assets: targetassets.Cluster,
	}

	targets = []target{installConfigTarget, manifestsTarget, ignitionConfigsTarget, clusterTarget, singleNodeIgnitionConfigTarget}
)

// clusterCreateError defines a custom error type that would help identify where the error occurs
// during the bootstrap phase of the installation process. This would help identify whether the error
// comes either from the Kubernetes API failure, the bootstrap failure or a general kubernetes client
// creation error. In the event of any error, this interface packages the error message and a custom
// log message that must be neatly presented to the user before termination of the project.
type clusterCreateError struct {
	wrappedError error
	logMessage   string
}

// Unwrap provides the actual stored error that occured during installation.
func (ce *clusterCreateError) Unwrap() error {
	return ce.wrappedError
}

// Error provides the actual stored error that occured during installation.
func (ce *clusterCreateError) Error() string {
	return ce.logMessage
}

// newAPIError creates a clusterCreateError object with a default error message specific to the API failure.
func newAPIError(errorInfo error) *clusterCreateError {
	return &clusterCreateError{
		wrappedError: errorInfo,
		logMessage: "Failed waiting for Kubernetes API. This error usually happens when there " +
			"is a problem on the bootstrap host that prevents creating a temporary control plane.",
	}
}

// newBootstrapError creates a clusterCreateError object with a default error message specific to the
// bootstrap failure.
func newBootstrapError(errorInfo error) *clusterCreateError {
	return &clusterCreateError{
		wrappedError: errorInfo,
		logMessage: "Failed to wait for bootstrapping to complete. This error usually " +
			"happens when there is a problem with control plane hosts that prevents " +
			"the control plane operators from creating the control plane.",
	}
}

// newClientError creates a clusterCreateError object with a default error message specific to the
// kubernetes client creation failure.
func newClientError(errorInfo error) *clusterCreateError {
	return &clusterCreateError{
		wrappedError: errorInfo,
		logMessage:   "Failed to create a kubernetes client.",
	}
}

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

// logComplete prints info upon completion
func logComplete(directory, consoleURL string) error {
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	kubeconfig := filepath.Join(absDir, "auth", "kubeconfig")
	pwFile := filepath.Join(absDir, "auth", "kubeadmin-password")
	pw, err := ioutil.ReadFile(pwFile)
	if err != nil {
		return err
	}
	logrus.Info("Install complete!")
	logrus.Infof("To access the cluster as the system:admin user when using 'oc', run 'export KUBECONFIG=%s'", kubeconfig)
	if consoleURL != "" {
		logrus.Infof("Access the OpenShift web-console here: %s", consoleURL)
		logrus.Infof("Login to the console with user: %q, and password: %q", "kubeadmin", pw)
	}
	return nil
}

func waitForInstallComplete(ctx context.Context, config *rest.Config, directory string) error {
	//if err := waitForInitializedCluster(ctx, config); err != nil {
	//		return err
	//	}

	//	consoleURL, err := getConsole(ctx, config)
	//	if err == nil {
	//		if err = addRouterCAToClusterCA(ctx, config, rootOpts.dir); err != nil {
	//			return err
	//		}
	//	} else {
	//		logrus.Warnf("Cluster does not have a console available: %v", err)
	//	}

	return logComplete(rootOpts.dir, "consoleURL")
}
