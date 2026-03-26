// backend/cronjob/service_test.go
package cronjob

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListCronJobs(t *testing.T) {
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cronjob",
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 * * * *",
		},
	}

	client := fake.NewSimpleClientset(cj)
	svc := NewService(client)

	jobs, err := svc.ListCronJobs(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 cronjob, got %d", len(jobs))
	}
	if jobs[0].Name != "my-cronjob" {
		t.Errorf("expected name 'my-cronjob', got %q", jobs[0].Name)
	}
}

func TestSuspendCronJob(t *testing.T) {
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cronjob",
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 * * * *",
		},
	}

	client := fake.NewSimpleClientset(cj)
	svc := NewService(client)

	err := svc.SetSuspend(context.Background(), "default", "my-cronjob", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := client.BatchV1().CronJobs("default").Get(context.Background(), "my-cronjob", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected error getting cronjob: %v", err)
	}
	if updated.Spec.Suspend == nil || !*updated.Spec.Suspend {
		t.Errorf("expected Suspend to be true, got %v", updated.Spec.Suspend)
	}
}

func TestTriggerJob(t *testing.T) {
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cronjob",
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "worker",
									Image: "busybox",
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(cj)
	svc := NewService(client)

	job, err := svc.TriggerJob(context.Background(), "default", "my-cronjob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected non-nil job")
	}
}
