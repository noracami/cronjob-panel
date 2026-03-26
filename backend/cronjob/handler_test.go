package cronjob

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/gin-gonic/gin"
)

func TestHandleListCronJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{Name: "cj1", Namespace: "default"},
		Spec:       batchv1.CronJobSpec{Schedule: "*/5 * * * *"},
	})

	r := gin.New()
	h := NewHandler(func(clusterID string) (*Service, error) {
		return NewService(client), nil
	})
	r.GET("/api/clusters/:id/cronjobs", h.ListCronJobs)

	req := httptest.NewRequest("GET", "/api/clusters/c1/cronjobs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}

	var result []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 1 {
		t.Fatalf("got %d items, want 1", len(result))
	}
}
