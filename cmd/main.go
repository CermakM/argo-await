package main

import (
	"fmt"

	"github.com/cermakm/argo-await/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	gjson "github.com/tidwall/gjson"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"os"
	"path/filepath"
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

// ENV VARS
const (
	// EnvVarKubeConfig is the path to the Kubernetes configuration
	EnvVarKubeConfig = "KUBE_CONFIG"

	// EnvVarDebugLog is the env var to turn on the debug mode for logging
	EnvVarDebugLog = "DEBUG_LOG"
)

var (
	res = &metav1.APIResource{
		Name:    "images",
		Group:   "image.openshift.io",
		Version: "v1",
		Kind:    "Image",
	} // FIXME

	namespace string
	filters   []string
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat:  "2019-08-08 12:00:00",
		FullTimestamp:    true,
		ForceColors:      true,
		QuoteEmptyFields: true,
	})

	flag.StringVarP(&namespace, "namespace", "n", "", "namespace to watch")
	flag.StringSliceVarP(&filters, "filter", "f", []string{}, "resource filter to be passed to jq processor")

	flag.Parse()
}

func main() {
	// FIXME: Debug
	var kubeConfig = filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	// kubeConfig, _ := os.LookupEnv(os.EnvVarKubeConfig)

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	config.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   res.Group,
		Version: res.Version,
	}

	if err != nil {
		log.Panic(err)
	}
	log.WithField("kubeconfig", kubeConfig).Debugf("Kubernetes config loaded.")

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

	log.WithFields(log.Fields{
		"group":   res.Group,
		"version": gvr.Version,
		"kind":    res.Kind,
	}).Info("watching for resources")
	for {
		select {
		case item := <-watchInterface.ResultChan():
			itemJSON := common.FormatJSON(item)

			contextLogger := log.WithFields(log.Fields{
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
