package observer

import (
	"errors"
	"fmt"

	"github.com/cermakm/argo-await/common"
	"github.com/sirupsen/logrus"
	gjson "github.com/tidwall/gjson"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/watch"
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

// AwaitResource awaits the given resource based on given filters
func (cl *AwaitResourceClient) AwaitResource(res *metav1.APIResource, namespace string, filters []string) {
	watchInterface, err := cl.Namespace(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		cl.Log.Panic(err)
	}

	cl.Log.WithFields(logrus.Fields{
		"group":   res.Group,
		"version": res.Version,
		"kind":    res.Kind,
	}).Info("watching for resources")
	for {
		select {
		case item := <-watchInterface.ResultChan():
			itemJSON := common.FormatJSON(item)

			contextLogger := cl.Log.WithFields(logrus.Fields{
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
