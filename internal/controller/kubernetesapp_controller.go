package controller

import (
    "context"
    "errors" // Leave this if used
    "github.com/go-logr/logr"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "k8s.io/apimachinery/pkg/api/errors"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    "sigs.k8s.io/controller-runtime/pkg/client"
    webappv1 "github.com/asfarahmad12/kubernetes-app-operator/api/v1alpha1" // Ensure this path is correct
)
// KubernetesAppReconciler reconciles a KubernetesApp object
type KubernetesAppReconciler struct {
    client.Client
    Log    logr.Logger
    Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KubernetesApp object
// and makes changes based on the state read and what is in the KubernetesApp.Spec
func (r *KubernetesAppReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
    ctx := context.Background()
    log := r.Log.WithValues("kubernetesapp", req.NamespacedName)

    // Fetch the KubernetesApp instance
    var app webappv1.KubernetesApp
    if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
        log.Error(err, "Unable to fetch KubernetesApp")
        return reconcile.Result{}, client.IgnoreNotFound(err)
    }

    // Define the desired state for the Deployment
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      app.Name,
            Namespace: app.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &app.Spec.Replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{"app": app.Name},
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{"app": app.Name},
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "kubernetes-app",
                            Image: app.Spec.Image,
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: int32(app.Spec.Port),
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    // Set the owner reference
    if err := controllerutil.SetControllerReference(&app, deployment, r.Scheme); err != nil {
        return reconcile.Result{}, err
    }

    // Create or update the Deployment
    found := &appsv1.Deployment{}
    err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
    if err != nil && errors.IsNotFound(err) {
        log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
        err = r.Create(ctx, deployment)
        if err != nil {
            return reconcile.Result{}, err
        }
    } else if err == nil {
        log.Info("Updating the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
        err = r.Update(ctx, deployment)
        if err != nil {
            return reconcile.Result{}, err
        }
    }

    return reconcile.Result{}, nil
}








