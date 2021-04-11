package controller

import (
	"fmt"

	"github.com/AdheipSingh/image-clone-controller/utils"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ignoreNamespacePredicate(), called before re-concilation loop in watcher
func IgnoreNamespacePredicate() predicate.Predicate {
	namespaces := utils.GetEnvAsSlice("DENY_LIST", nil, ",")
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			for _, namespace := range namespaces {
				if e.Object.GetNamespace() == namespace {
					msg := fmt.Sprintf("controller will not re-concile namespace [%s], alter DENY_LIST to re-concile", e.Object.GetNamespace())
					log.Log.Info(msg)
					return false
				}
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			for _, namespace := range namespaces {
				if e.ObjectNew.GetNamespace() == namespace {
					msg := fmt.Sprintf("controller will not re-concile namespace [%s], alter DENY_LIST to re-concile", e.ObjectNew.GetNamespace())
					log.Log.Info(msg)
					return false
				}
			}
			return true
		},
	}
}
