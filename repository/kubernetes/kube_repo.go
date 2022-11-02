package kubernetes

import (
	"context"
	"encoding/json"

	"github.com/panagiotisptr/job-scheduler/repository"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyv1 "k8s.io/client-go/applyconfigurations/batch/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesRepository struct {
	logger *zap.Logger
	client *kubernetes.Clientset
}

func ProvideKubernetesRepository(
	logger *zap.Logger,
	client *kubernetes.Clientset,
) (repository.KubernetesRepository, error) {
	repo := &KubernetesRepository{
		logger: logger,
		client: client,
	}

	return repo, nil
}

func (r *KubernetesRepository) GetRunningJobs(
	ctx context.Context,
) ([]batchv1.CronJob, error) {
	cronJobs, err := r.client.BatchV1().CronJobs("").List(
		ctx,
		metav1.ListOptions{},
	)
	if err != nil {
		return []batchv1.CronJob{}, err
	}

	return cronJobs.Items, nil
}

// This is sketchy
func cronJobToConfig(
	cj batchv1.CronJob,
) (applyv1.CronJobApplyConfiguration, error) {
	config := applyv1.CronJobApplyConfiguration{}

	b, err := json.Marshal(cj)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(b, &config)

	return config, err
}

func (r *KubernetesRepository) StartJob(
	ctx context.Context,
	cj batchv1.CronJob,
) error {
	_, err := r.client.BatchV1().CronJobs("default").Create(
		ctx,
		&cj,
		metav1.CreateOptions{},
	)

	return err
}

func (r *KubernetesRepository) StopJob(
	ctx context.Context,
	cj batchv1.CronJob,
) error {
	c, err := r.client.BatchV1().CronJobs("default").Get(
		ctx,
		cj.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	t := true
	c.Spec.Suspend = &t

	_, err = r.client.BatchV1().CronJobs("default").Update(
		ctx,
		c,
		metav1.UpdateOptions{},
	)

	return err
}
