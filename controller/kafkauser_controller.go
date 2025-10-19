package controller

import (
	"context"
	"fmt"

	v1 "github.com/crafter76/demo-kafka-operator/api/v1"
	"github.com/crafter76/demo-kafka-operator/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const finalizerName = "kafka.appfarm.rs/finalizer"

type KafkaUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Kafka  *mocks.MockKafkaClient
}

func (r *KafkaUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconcile started", "name", req.Name, "namespace", req.Namespace)

	var kafkaUser v1.KafkaUser
	if err := r.Get(ctx, req.NamespacedName, &kafkaUser); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	original := kafkaUser.DeepCopy()

	if !kafkaUser.DeletionTimestamp.IsZero() {
		if contains(kafkaUser.Finalizers, finalizerName) {
			r.Kafka.DeleteUser(req.Name)
			log.Info("Deleted user from Kafka", "username", req.Name)

			kafkaUser.Finalizers = remove(kafkaUser.Finalizers, finalizerName)
			if err := r.Patch(ctx, &kafkaUser, client.MergeFrom(original)); err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
			log.Info("Finalizer removed")
		}
		return ctrl.Result{}, nil
	}

	if !contains(kafkaUser.Finalizers, finalizerName) {
		log.Info("Adding finalizer")
		kafkaUser.Finalizers = append(kafkaUser.Finalizers, finalizerName)
		if err := r.Patch(ctx, &kafkaUser, client.MergeFrom(original)); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if r.Kafka.HasUser(req.Name) {
		log.Info("User already exists in Kafka, skipping creation")
	} else {
		log.Info("Creating user in Kafka via mock")
		if err := r.Kafka.CreateUser(req.Name); err != nil {
			log.Error(err, "Failed to create user in Kafka")
			updateCondition(&kafkaUser, "Created", metav1.ConditionFalse, "CreateFailed", err.Error())
			if statusErr := r.Status().Patch(ctx, &kafkaUser, client.MergeFrom(original)); statusErr != nil {
				log.Error(statusErr, "Failed to update status after error")
			}
			return ctrl.Result{}, fmt.Errorf("failed to create user in kafka: %w", err)
		}
		log.Info("User created in Kafka")
	}

	log.Info("Updating status: Created=True")
	updateCondition(&kafkaUser, "Created", metav1.ConditionTrue, "Success", "User created in Kafka")

	if err := r.Status().Patch(ctx, &kafkaUser, client.MergeFrom(original)); err != nil {
		log.Error(err, "Failed to update status, will retry")
		return ctrl.Result{Requeue: true}, nil
	}

	log.Info("Reconcile completed successfully")
	return ctrl.Result{}, nil
}

func updateCondition(cr *v1.KafkaUser, conditionType string, status metav1.ConditionStatus, reason, msg string) {
	idx := -1
	for i, cond := range cr.Status.Conditions {
		if cond.Type == conditionType {
			idx = i
			break
		}
	}

	newCond := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            msg,
	}

	if idx == -1 {
		cr.Status.Conditions = append(cr.Status.Conditions, newCond)
	} else {
		cr.Status.Conditions[idx] = newCond
	}
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	result := make([]string, 0)
	for _, item := range list {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func (r *KafkaUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.KafkaUser{}).
		Complete(r)
}
