package resultsservice

import (
	"bytes"
	"encoding/json"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"io"
	"net/http"
	"testing"
)

func TestResultsService(t *testing.T) {
	g := NewGomegaWithT(t)
	GetTaskRun = func(service *ResultsService, namespace string, taskRun string) (*v1.TaskRun, error) {
		g.Expect(namespace).Should(Equal("default"))
		g.Expect(taskRun).Should(Equal("task1"))
		return &v1.TaskRun{
			Status: v1.TaskRunStatus{},
		}, nil
	}
	UpdateTaskRun = func(service *ResultsService, tr *v1.TaskRun) error {
		g.Expect(tr.Status.Results[0].Name).Should(Equal("foo"))
		g.Expect(tr.Status.Results[0].Value.StringVal).Should(Equal("bar"))
		return nil
	}

	rm := ResultsMessage{
		TaskRunName: "task1",
		Namespace:   "default",
		Results:     map[string]string{"foo": "bar"},
	}
	data, _ := json.Marshal(&rm)

	rs := ResultsService{Logger: &logr.Logger{}}
	responseWriter := &TestHttpResponse{}
	rs.ServeHTTP(responseWriter, &http.Request{
		Body: io.NopCloser(bytes.NewReader(data)),
	})
	g.Expect(responseWriter.statusCode).Should(Equal(204))
}

type TestHttpResponse struct {
	statusCode int
}

func (t *TestHttpResponse) Header() http.Header {
	//TODO implement me
	panic("implement me")
}

func (t *TestHttpResponse) Write(i []byte) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TestHttpResponse) WriteHeader(statusCode int) {
	t.statusCode = statusCode
}
