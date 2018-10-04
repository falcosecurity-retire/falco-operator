package app

import (
	"github.com/sirupsen/logrus"
	"runtime"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"time"
	"context"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	Resource = "mumoshu.github.io/v1alpha1"
	Kind     = "FalcoRule"
)

type OperateOpts struct {
	ConfigMapName      string
	ConfigMapNamespace string
	WatchNamespace     string
}

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func Operate(opts OperateOpts) error {
	printVersion()

	sdk.ExposeMetricsPort()

	namespace := opts.WatchNamespace

	//namespace, err := k8sutil.GetWatchNamespace()
	//if err != nil {
	//	logrus.Fatalf("failed to get watch namespace: %v", err)
	//	logrus.Infof("falling back to all namespaces")
	//	namespace = ""
	//}
	resyncPeriod := time.Duration(5) * time.Second
	logrus.Infof("Watching %s, %s, %s, %d", Resource, Kind, namespace, resyncPeriod)
	sdk.Watch(Resource, Kind, namespace, resyncPeriod)
	sdk.Handle(NewHandler(opts))
	sdk.Run(context.TODO())

	return nil
}
