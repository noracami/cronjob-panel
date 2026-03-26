// backend/cronjob/service.go
package cronjob

import (
	"context"
	"fmt"
	"io"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
)

type Service struct {
	client kubernetes.Interface
}

func NewService(client kubernetes.Interface) *Service {
	return &Service{client: client}
}

func (s *Service) ListCronJobs(ctx context.Context, namespace string) ([]batchv1.CronJob, error) {
	list, err := s.client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (s *Service) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	return s.client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (s *Service) ListJobs(ctx context.Context, namespace, cronjobName string) ([]batchv1.Job, error) {
	list, err := s.client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var jobs []batchv1.Job
	for _, job := range list.Items {
		for _, ref := range job.OwnerReferences {
			if ref.Kind == "CronJob" && ref.Name == cronjobName {
				jobs = append(jobs, job)
				break
			}
		}
	}
	return jobs, nil
}

func (s *Service) GetPodLog(ctx context.Context, namespace, podName string) (string, error) {
	req := s.client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: ptr.To(int64(500)),
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	bytes, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Service) TriggerJob(ctx context.Context, namespace, cronjobName string) (*batchv1.Job, error) {
	cj, err := s.GetCronJob(ctx, namespace, cronjobName)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-manual-%d", cronjobName, time.Now().Unix()),
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
					Name:       cj.Name,
					UID:        cj.UID,
				},
			},
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}

	return s.client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
}

func (s *Service) SetSuspend(ctx context.Context, namespace, name string, suspend bool) error {
	cj, err := s.GetCronJob(ctx, namespace, name)
	if err != nil {
		return err
	}
	cj.Spec.Suspend = ptr.To(suspend)
	_, err = s.client.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
	return err
}
