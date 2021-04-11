package controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"

	"github.com/AdheipSingh/image-clone-controller/utils"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var Auth = authn.Basic{Username: os.Getenv("USERNAME"), Password: os.Getenv("PASSWORD")}

// reconcileImageController reconciles reconcileImageController
type ReconcileImageController struct {
	Client client.Client
}

var _ reconcile.Reconciler = &ReconcileImageController{}

func (r *ReconcileImageController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	if request.Namespace == "kube-system" {
		return reconcile.Result{Requeue: false}, nil
	}

	//----------------------------------- MAKE DEPLOYMENT LIST --------------------------------------
	deploymentListEmptyObject := &appsv1.DeploymentList{}
	err := r.Client.List(ctx, deploymentListEmptyObject)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not List Deployments")
		return reconcile.Result{}, nil
	}

	// create a slice []*deploymentList
	deploymentList := make([]*v1.Deployment, 0)
	for i := range deploymentListEmptyObject.Items {
		deploymentList = append(deploymentList, &deploymentListEmptyObject.Items[i])
	}

	//----------------------------------------- MAKE DAEMONSET LIST ----------------------------------------

	daemonListEmptyObject := &appsv1.DaemonSetList{}
	err = r.Client.List(ctx, daemonListEmptyObject)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not List Daemonsets")
		return reconcile.Result{}, nil
	}

	// create a slice []*deploymentList
	daemonList := make([]*v1.DaemonSet, 0)
	for i := range daemonListEmptyObject.Items {
		daemonList = append(daemonList, &daemonListEmptyObject.Items[i])
	}

	images, err := getImageFromBackUpRegistery()
	if err != nil {
		return reconcile.Result{}, nil
	}

	for i := range deploymentList {
		// dont do anything on controller
		if deploymentList[i].Name == "image-control-controller" {
			return reconcile.Result{}, nil
		} else {
			fmt.Println("aaaa")
			if deploymentList[i].Annotations["backup"] != "true" {

				if !utils.ContainsString(images, strings.Replace(deploymentList[i].Spec.Template.Spec.Containers[i].Image, ":", "", -1)) {
					log.Info("Image does not exist, in backupregistery, taking backup of registery", "imageName", deploymentList[i].Spec.Template.Spec.Containers[i].Image)
					tag, err := pullObjectImageTagandPush(deploymentList[i].Spec.Template.Spec.Containers[i].Image)
					if err != nil {
						log.Error(err, err.Error())
						return reconcile.Result{}, nil
					}
					patch := client.MergeFrom(deploymentList[i].DeepCopy())
					deploymentList[i].Annotations = map[string]string{
						"backup": "true",
					}
					deploymentList[i].Spec.Template.Spec.Containers[i].Image = tag
					err = r.Client.Patch(context.TODO(), deploymentList[i], patch)
					if err != nil {
						log.Error(err, err.Error())
						return reconcile.Result{}, nil
					} else {
						log.Info("Patched Success", "deployment", deploymentList[i].Name)
					}
				}
			}
		}

	}

	for i := range daemonList {
		if daemonList[i].Annotations["backup"] != "true" {

			if !utils.ContainsString(images, strings.Replace(daemonList[i].Spec.Template.Spec.Containers[i].Image, ":", "", -1)) {
				log.Info("Image does not exist, in backupregistery, taking backup of registery", "imageName", daemonList[i].Spec.Template.Spec.Containers[i].Image)
				tag, err := pullObjectImageTagandPush(daemonList[i].Spec.Template.Spec.Containers[i].Image)
				if err != nil {
					log.Error(err, err.Error())
					return reconcile.Result{}, nil
				}
				patch := client.MergeFrom(daemonList[i].DeepCopy())
				daemonList[i].Annotations = map[string]string{
					"backup": "true",
				}
				daemonList[i].Spec.Template.Spec.Containers[i].Image = tag
				err = r.Client.Patch(context.TODO(), daemonList[i], patch)
				if err != nil {
					log.Error(err, err.Error())
					return reconcile.Result{}, nil
				} else {
					log.Info("Patched Success", "daemonset", daemonList[i].Name)
				}
			}
		}

	}

	return reconcile.Result{}, nil
}

func getImageFromBackUpRegistery() ([]string, error) {

	ref, err := name.NewRepository(os.Getenv("REGISTERY"))
	if err != nil {
		return nil, err
	}

	img, err := remote.List(ref, remote.WithAuth(&Auth))
	if err != nil {
		return nil, err
	}

	return img, nil

}

func pullObjectImageTagandPush(imageName string) (newImageName string, err error) {

	image, err := crane.Pull(imageName, crane.WithAuth(&Auth))
	if err != nil {
		return "", err
	}

	igN := strings.Replace(imageName, ":", "", -1)
	tag, err := name.NewTag(os.Getenv("REGISTERY") + igN)
	if err != nil {
		return "", err
	}

	//	fmt.Println(tag.String())
	if err := crane.Push(image, tag.String(), crane.WithAuth(&Auth)); err != nil {
		return "", err
	}

	return tag.String(), nil
}
