package main

import (
	goflag "flag"
	"fmt"
	"k8s.io/klog"

	"github.com/cermakm/argo-await/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	gjson "github.com/tidwall/gjson"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"os"
)

func passFilters(evt *watch.Event, filters ...string) (bool, error) {
	eventJSON := common.FormatJSON([]watch.Event{*evt})

	if !gjson.Valid(eventJSON) {
		return false, errors.New("failed parsing event: invalid json")
	}

	for _, f := range filters {
		// filter needs to wrapped
		wrappedFilter := fmt.Sprintf("#(%s)", f)
		validResource := gjson.Get(eventJSON, wrappedFilter)

		if !validResource.Exists() {
			return false, nil
		}
	}

	return true, nil
}

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
		ns, _, err := clientcmd.DefaultClientConfig.Namespace()
		if err != nil {
			// cannot proceed with empty namespace
			log.Panic("Namespace was not provided and could not be determined.")
		}
		log.WithField("namespace", ns).Info("Setting default namespace.")
		namespace = ns
	}

	dynamicClient := dynamic.NewForConfigOrDie(config)

	gvr := schema.GroupVersionResource{
		Group:    res.Group,
		Version:  res.Version, // apiResource seems to be have empty Version string
		Resource: res.Name,
	}
	resourceClient := dynamicClient.Resource(gvr)

	watchInterface, err := resourceClient.Namespace(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		log.Panic(err)
	}

	log.WithFields(logrus.Fields{
		"group":   res.Group,
		"version": gvr.Version,
		"kind":    res.Kind,
	}).Info("watching for resources")
	for {
		select {
		case item := <-watchInterface.ResultChan():
			itemJSON := common.FormatJSON(item)

			contextLogger := log.WithFields(logrus.Fields{
				"type":     item.Type,
				"resource": item.Object.GetObjectKind().GroupVersionKind(),
			})
			contextLogger.Info("New resource received")
			contextLogger.Debugf("Data: %#v", itemJSON)

			gvk := item.Object.GetObjectKind().GroupVersionKind()
			if res.Kind != gvk.Kind {
				contextLogger.Infof("resource does not match required kind: '%s'", res.Kind)
				continue
			}

			contextLogger.Infof("applying filter: %#v", filters)

			if ok, err := passFilters(&item, filters...); ok == true {
				contextLogger.Info("resource fulfilled")
				return
			} else if err != nil {
				contextLogger.Error(err)
			}

			contextLogger.Info("resource dit not pass the filter")
		}
	}

}
