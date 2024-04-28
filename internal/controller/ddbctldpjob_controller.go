/*
Copyright 2024.

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

package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dynamoctlv1alpha1 "github.com/Sanjiv-Madhavan/dynamoctl-dp-operator/api/v1alpha1"
)

var (
	jobOwnerKey = ".metadata.controller"
	apiGVString = dynamoctlv1alpha1.GroupVersion.String()
)

// DdbctlDpJobReconciler reconciles a DdbctlDpJob object
type DdbctlDpJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dynamoctl.dp.operators.sanjivmadhavan.io,resources=ddbctldpjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dynamoctl.dp.operators.sanjivmadhavan.io,resources=ddbctldpjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dynamoctl.dp.operators.sanjivmadhavan.io,resources=ddbctldpjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DdbctlDpJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *DdbctlDpJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("DynamoDB Delete Partition Reconciliation", "status", "started")

	var ddbctldpjob dynamoctlv1alpha1.DdbctlDpJob
	if err := r.Get(ctx, req.NamespacedName, &ddbctldpjob); err != nil {
		log.Error(err, "unable to fetch DeleteTablePartitionJob")
		return ctrl.Result{}, err
	}

	// Create Pod Spec
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "ddbctl-delete-partition",
				Image:   "sanjivmadhavan/go-dynamodb-partition-delete:latest",
				Command: []string{"/dynamoctl", "delete-partition"},
				Args: []string{
					"-t",
					ddbctldpjob.Spec.TableName,
					"-p",
					ddbctldpjob.Spec.PartitionValue,
					"-e",
					ddbctldpjob.Spec.EndpointURL,
					"-r",
					ddbctldpjob.Spec.AWSRegion,
					"-s",
				},
			},
		},
		RestartPolicy: corev1.RestartPolicyNever,
	}

	jobTemplate := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ddbctldpjob.Name + "-job",
			Namespace: req.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "ddbctl-dp"},
				},
				Spec: podSpec,
			},
		},
	}

	// Set DeleteTablePartitionJob instance as a the owner of the Job
	if err := ctrl.SetControllerReference(&ddbctldpjob, jobTemplate, r.Scheme); err != nil {
		log.Error(err, "unable to set controller reference (jobOwner) for the Job")
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, &ddbctldpjob); err != nil {
		log.Error(err, "unable to create Job for DeleteTablePartitionJob")
		return ctrl.Result{}, err
	}

	log.Info("DynamoDB Delete Partition Reconciliation", "status", "completed")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DdbctlDpJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Get fieldIndexer and add the field jobOwner
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		job := rawObj.(*batchv1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVString || owner.Kind != "DdbctlDpJob" {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&dynamoctlv1alpha1.DdbctlDpJob{}).
		Owns(&batchv1.Job{}).
		Complete(r)
	// creates a controller, adds the reconciler to the controller, binds it to manager using builder.
}
