package app

import (
	"github.com/sirupsen/logrus"
	"runtime"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"time"
	"github.com/mumoshu/falco-operator/pkg/stub"
	"context"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	Resource = "mumoshu.github.io/v1alpha1"
	Kind     = "FalcoRule"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func Operate() error {
	printVersion()

	sdk.ExposeMetricsPort()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("failed to get watch namespace: %v", err)
		return err
	}
	resyncPeriod := time.Duration(5) * time.Second
	logrus.Infof("Watching %s, %s, %s, %d", Resource, Kind, namespace, resyncPeriod)
	sdk.Watch(Resource, Kind, namespace, resyncPeriod)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())

	return nil
}
