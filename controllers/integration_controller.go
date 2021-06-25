/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"

	uniformv1alpha1 "github.com/bacherfl/keptn-uniform-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IntegrationReconciler reconciles a Integration object
type IntegrationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=uniform.my.domain,resources=integrations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=uniform.my.domain,resources=integrations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=uniform.my.domain,resources=integrations/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Integration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *IntegrationReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	//reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	//reqLogger.Info("Reconciling Uniform")

	// Fetch the Uniform instance
	instance := &uniformv1alpha1.Integration{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	retriggerReconcile := false

	deplForCR, err := newDeploymentForCR(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	// Set Uniform instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, deplForCR, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the desired Deployment already exists
	foundDepl := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deplForCR.Name, Namespace: deplForCR.Namespace}, foundDepl)
	if err != nil && errors.IsNotFound(err) {
		//reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deplForCR.Namespace, "Deployment.Name", deplForCR.Name)
		err = r.Create(context.TODO(), deplForCR)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Pod created successfully - don't requeue
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if strings.Contains(foundDepl.Spec.Template.Labels["app.kubernetes.io/managed-by"], "skaffold") {
		//reqLogger.Info("Not upgrading resource managed by skaffold", "Deployment.Namespace", foundDepl.Namespace, "Deployment.Name", foundDepl.Name)
		// if the pod is managed by skaffold (i.e., a dev is currently debugging it, do not overwrite the pod
		retriggerReconcile = true
		return reconcile.Result{}, err
	}
	foundDepl.Spec = deplForCR.Spec

	err = r.Update(context.TODO(), foundDepl)

	if err != nil {
		//reqLogger.Error(err, "Could not update service")
		return reconcile.Result{}, err
	}

	// Deployment already exists - don't requeue
	//reqLogger.Info("Updated deployment", "Deployment.Namespace", foundDepl.Namespace, "Deployment.Name", foundDepl.Name)

	serviceForCR := newServiceForCR(instance)

	if err := controllerutil.SetControllerReference(instance, serviceForCR, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundSvc := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: serviceForCR.Name, Namespace: serviceForCR.Namespace}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		//reqLogger.Info("Creating a new Service", "Service.Namespace", serviceForCR.Namespace, "Service.Name", serviceForCR.Name)
		err = r.Create(context.TODO(), serviceForCR)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Service already exists - don't requeue
	//reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", foundSvc.Namespace, "Service.Name", foundSvc.Name)

	if retriggerReconcile {
		// retrigger reconciliation
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IntegrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&uniformv1alpha1.Integration{}).
		Complete(r)
}

// newDeploymentForCR returns a deployment with the same name/namespace as the cr
func newDeploymentForCR(cr *uniformv1alpha1.Integration) (*appsv1.Deployment, error) {

	f := func(s int32) *int32 {
		return &s
	}

	envVars := cr.Spec.Env

	version := "n/a"
	split := strings.Split(cr.Spec.Image, ":")
	if len(split) > 1 {
		version = split[1]
	}

	integrationDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: "keptn",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: f(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": cr.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"run":                       cr.Name,
						"app.kubernetes.io/version": version,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cr.Name,
							Image: cr.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: envVars,
						},
						{
							Name:  "distributor",
							Image: "keptn/distributor:0.8.4",
							SecurityContext: &corev1.SecurityContext{
								RunAsUser:                intp(65532),
								RunAsNonRoot:             boolp(true),
								ReadOnlyRootFilesystem:   boolp(true),
								AllowPrivilegeEscalation: boolp(false),
							},
						},
					},
				},
			},
		},
	}

	distributorEnvVars, err := getDistributorEnvVars(cr)
	if err != nil {
		return nil, nil
	}

	integrationDeployment.Spec.Template.Spec.Containers[1].Env = distributorEnvVars

	return integrationDeployment, nil
}

func getDistributorEnvVars(cr *uniformv1alpha1.Integration) ([]corev1.EnvVar, error) {
	topics := ""

	for i, topic := range cr.Spec.Events {
		if cr.Spec.Remote != nil && strings.ContainsAny(topic, "*>") {
			return nil, errors.NewBadRequest("cannot subscribe to wildcard topics in remote execution plane mode")
		}
		topics = topics + topic
		if i < len(cr.Spec.Events)-1 {
			topics = topics + ","
		}
	}

	envVars := []corev1.EnvVar{
		{
			Name:  "PUBSUB_RECIPIENT",
			Value: cr.Name,
		},
		{
			Name:  "PUBSUB_TOPIC",
			Value: topics,
		},
		{
			Name: "VERSION",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.labels['app.kubernetes.io/version']",
				},
			},
		},
		{
			Name:  "K8S_DEPLOYMENT_NAME",
			Value: cr.Name,
		},
		{
			Name: "K8S_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "K8S_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "K8S_NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "spec.nodeName",
				},
			},
		},
	}

	if cr.Spec.Remote != nil {
		var err error
		envVars, err = appendRemoteExecutionEnvVars(envVars, cr.Spec.Remote)
		if err != nil {
			return nil, nil
		}
	} else {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "PUBSUB_URL",
			Value: "nats://keptn-nats-cluster",
		})
	}

	return envVars, nil
}

func appendRemoteExecutionEnvVars(vars []corev1.EnvVar, remote *uniformv1alpha1.RemoteExecutionPlaneSpec) ([]corev1.EnvVar, error) {
	remoteEnvVars := []corev1.EnvVar{
		{
			Name:  "KEPTN_API_ENDPOINT",
			Value: remote.APIURL,
		},
	}

	var apiTokenEnvVar corev1.EnvVar
	if remote.APIToken.Value != "" && remote.APIToken.FromSecret == nil {
		apiTokenEnvVar = corev1.EnvVar{
			Name:  "KEPTN_API_TOKEN",
			Value: remote.APIToken.Value,
		}
	} else if remote.APIToken.Value == "" && remote.APIToken.FromSecret != nil {
		apiTokenEnvVar = corev1.EnvVar{
			Name: "KEPTN_API_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: remote.APIToken.FromSecret.Field,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: remote.APIToken.FromSecret.SecretName,
					},
				},
			},
		}
	} else {
		return nil, errors.NewBadRequest("either spec.remote.apiToken.Value or spec.remote.apiToken.ValueFrom has to be set")
	}

	if remote.DisableSSLVerification {
		remoteEnvVars = append(remoteEnvVars, corev1.EnvVar{
			Name:  "HTTP_SSL_VERIFY",
			Value: "false",
		})
	}

	remoteEnvVars = append(remoteEnvVars, apiTokenEnvVar)

	vars = append(vars, remoteEnvVars...)
	return vars, nil
}

func boolp(b bool) *bool {
	return &b
}

func intp(i int64) *int64 {
	return &i
}

func ensureEnvVarIsSet(name string, value string, envVars []corev1.EnvVar) []corev1.EnvVar {

	for _, env := range envVars {
		if env.Name == name {
			return envVars
		}
	}

	newEnvVar := corev1.EnvVar{
		Name:  name,
		Value: value,
	}

	envVars = append(envVars, newEnvVar)
	return envVars
}

func newServiceForCR(cr *uniformv1alpha1.Integration) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: "keptn",
			Labels: map[string]string{
				"run": cr.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol: "TCP",
					Port:     8080,
				},
			},
			Selector: map[string]string{
				"run": cr.Name,
			},
		},
		Status: corev1.ServiceStatus{},
	}
}
