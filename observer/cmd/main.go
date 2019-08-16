package main

import (
	goflag "flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/cermakm/argo-await/common"
	"github.com/cermakm/argo-await/observer"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	flag "github.com/spf13/pflag"

	"os"
)

var (
	log = common.Logger()

	res = &metav1.APIResource{
		Name:    "configmaps",
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	} // FIXME

	namespace string
	filters   []string
)

func init() {
	flag.StringVarP(&namespace, "namespace", "n", "", "namespace to watch")
	flag.StringSliceVarP(&filters, "filter", "f", []string{}, "resource filter to be passed to jq processor")

	// initialize klog flags
	flagset := goflag.CommandLine
	klog.InitFlags(flagset)

	flag.CommandLine.AddGoFlagSet(flagset)
	flag.CommandLine.SortFlags = false

	flag.Parse()
}

func main() {
	kubeConfig, _ := os.LookupEnv(common.EnvVarKubeConfig)

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.WithField("config", config).Panic(err)
	}
	config.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   res.Group,
		Version: res.Version,
	}
	log.WithField("kubeconfig", kubeConfig).Debugf("Kubernetes config loaded.")

	if namespace == "" {
		ns, err := k8sutil.GetOperatorNamespace()
		if err != nil {
			// cannot proceed with empty namespace
			log.Warnf("Cannot establish operator namespace. Attempting to read from config.")

			ns, _, err := clientcmd.DefaultClientConfig.Namespace()
			if err != nil {
				// cannot proceed with empty namespace
				log.Panic("Namespace was not provided and could not be determined.")
			}
			log.WithField("namespace", ns).Info("Setting default namespace.")
			namespace = ns
		}

		namespace = ns
	}

	client, err := observer.NewClientForConfigResource(config, res)
	client.AwaitResource(res, namespace, filters)
}
