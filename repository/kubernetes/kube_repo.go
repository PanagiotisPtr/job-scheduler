package kubernetes

import (
	"context"
	"os"

	"github.com/panagiotisptr/job-scheduler/repository"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *KubernetesRepository) getCronJob(
	ctx context.Context,
	cj *batchv1.CronJob,
	createIfNotFound bool,
) (*batchv1.CronJob, error) {
	cronJob, err := r.client.BatchV1().CronJobs(r.getNamespace()).Get(
		ctx,
		cj.Name,
		metav1.GetOptions{},
	)
	if err == nil {
		return cronJob, nil
	}

	// Basically if it is not a not found error we return the error
	// If it is a not found error we return it only if we don't want
	// to create the cron job
	if !errors.IsNotFound(err) || !createIfNotFound {
		return cronJob, err
	}

	return r.client.BatchV1().CronJobs(r.getNamespace()).Create(
		ctx,
		cj,
		metav1.CreateOptions{},
	)
}

func (r *KubernetesRepository) getNamespace() string {
	envNamespace := os.Getenv("POD_NAMESPACE")
	if envNamespace != "" {
		return envNamespace
	}

	return metav1.NamespaceDefault
}

func (r *KubernetesRepository) GetRunningCronJobs(
	ctx context.Context,
) ([]string, error) {
	names := []string{}
	cronJobs, err := r.client.BatchV1().CronJobs(r.getNamespace()).List(
		ctx,
		metav1.ListOptions{},
	)
	if err != nil {
		return names, err
	}
	for _, cj := range cronJobs.Items {
		if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
			continue
		}
		names = append(names, cj.Name)
	}

	return names, nil
}

func (r *KubernetesRepository) StartCronJob(
	ctx context.Context,
	cj *batchv1.CronJob,
) error {
	cronJob, err := r.getCronJob(ctx, cj, true)
	if err != nil {
		return err
	}

	isSuspended := func(c *batchv1.CronJob) bool {
		return c.Spec.Suspend != nil && *c.Spec.Suspend
	}

	// this cronjob was previously stopped and the spec says
	// it should be running. Delete and recreate it
	if isSuspended(cronJob) && !isSuspended(cj) {
		err = r.client.BatchV1().CronJobs(r.getNamespace()).Delete(
			ctx,
			cronJob.Name,
			metav1.DeleteOptions{},
		)
		if err != nil {
			return err
		}

		// This is safe - it won't get stuck in a loop
		// we just deleted the object so next time isSuspended
		// will match cj
		return r.StartCronJob(ctx, cj)
	}
	t := false
	cronJob.Spec.Suspend = &t

	_, err = r.client.BatchV1().CronJobs(r.getNamespace()).Update(
		ctx,
		cronJob,
		metav1.UpdateOptions{},
	)

	return err
}

func (r *KubernetesRepository) StopCronJob(
	ctx context.Context,
	cj *batchv1.CronJob,
) error {
	c, err := r.getCronJob(ctx, cj, false)
	if err != nil {
		return err
	}
	t := true
	c.Spec.Suspend = &t

	_, err = r.client.BatchV1().CronJobs(r.getNamespace()).Update(
		ctx,
		c,
		metav1.UpdateOptions{},
	)

	return err
}
