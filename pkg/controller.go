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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, Add)
}

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileImageController{Client: mgr.GetClient()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("image-clone-controller", mgr, controller.Options{
		Reconciler: &ReconcileImageController{Client: mgr.GetClient()},
	})
	if err != nil {
		return err
	}

	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}, IgnoreNamespacePredicate()); err != nil {
		return err
	}

	if err := c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForObject{}, IgnoreNamespacePredicate()); err != nil {
		return err
	}

	return nil
}

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

// reconcileImageController reconciles reconcileImageController
type ReconcileImageController struct {
	Client client.Client
}

var _ reconcile.Reconciler = &ReconcileImageController{}

func (r *ReconcileImageController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

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

	fmt.Println(images)
	for i := range deploymentList {
		if deploymentList[i].Annotations["backup"] == "true" {
			return reconcile.Result{}, nil
		} else {
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

	return reconcile.Result{}, nil
}

func getImageFromBackUpRegistery() ([]string, error) {

	ref, err := name.NewRepository("docker.io/imageclonecontroller/backup-registery")
	if err != nil {
		return nil, err
	}

	var a = authn.Basic{Username: os.Getenv("USERNAME"), Password: os.Getenv("PASSWORD")}
	// Fetch the manifest using default credentials.
	img, err := remote.List(ref, remote.WithAuth(&a))
	if err != nil {
		return nil, err
	}

	return img, nil

}

func pullObjectImageTagandPush(imageName string) (newImageName string, err error) {
	var a = authn.Basic{Username: os.Getenv("USERNAME"), Password: os.Getenv("PASSWORD")}

	image, err := crane.Pull(imageName, crane.WithAuth(&a))
	if err != nil {
		return "", err
	}

	igN := strings.Replace(imageName, ":", "", -1)
	tag, err := name.NewTag("docker.io/imageclonecontroller/backup-registery:" + igN)
	if err != nil {
		return "", err
	}

	//	fmt.Println(tag.String())
	if err := crane.Push(image, tag.String(), crane.WithAuth(&a)); err != nil {
		return "", err
	}

	return tag.String(), nil
}
