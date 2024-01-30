package resultsservice

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	pipelineclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

// service service example implementation.
// The example methods log the requests and return zero values.
type ResultsService struct {
	Logger *logr.Logger
	Client *kubernetes.Clientset
}

type ResultsMessage struct {
	TaskRunName string
	Namespace   string
	Results     map[string]string
}

var GetTaskRun = DefaultGetTaskRun
var UpdateTaskRun = DefaultUpdateTaskRun

func (s *ResultsService) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		s.Logger.Error(err, "failed to read request body", "address", request.RemoteAddr)
		writer.WriteHeader(500)
		return
	}

	message := ResultsMessage{}
	err = json.Unmarshal(body, &message)
	if err != nil {
		writer.WriteHeader(400)
		return
	}
	tr, err := GetTaskRun(s, message.Namespace, message.TaskRunName)

	if err != nil {
		writer.WriteHeader(500)
		return
	}
	for _, res := range tr.Status.Results {
		_, exists := message.Results[res.Name]
		if exists {
			// The results have already been set, this is not good
			// it means something funky is going on and we should remove all results, as they may be bad
			tr.Status.Results = []v1.TaskRunResult{}
			err = UpdateTaskRun(s, tr)
			if err != nil {
				writer.WriteHeader(500)
				return
			}
			writer.WriteHeader(400)
			return
		}
	}
	for key, value := range message.Results {
		if err != nil {
			writer.WriteHeader(500)
			return
		}
		tr.Status.Results = append(tr.Status.Results, v1.TaskRunResult{Name: key, Type: v1.ResultsTypeString, Value: v1.ResultValue{Type: v1.ParamTypeString, StringVal: value}})
	}
	err = UpdateTaskRun(s, tr)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	writer.WriteHeader(204)
}

func DefaultGetTaskRun(service *ResultsService, namespace string, taskRun string) (*v1.TaskRun, error) {
	tektonClient := pipelineclientset.New(service.Client.RESTClient())
	trClient := tektonClient.TektonV1().TaskRuns(namespace)
	tr, err := trClient.Get(context.Background(), taskRun, metav1.GetOptions{})
	return tr, err
}
func DefaultUpdateTaskRun(service *ResultsService, tr *v1.TaskRun) error {
	tektonClient := pipelineclientset.New(service.Client.RESTClient())
	trClient := tektonClient.TektonV1().TaskRuns(tr.Namespace)
	_, err := trClient.UpdateStatus(context.Background(), tr, metav1.UpdateOptions{})
	return err

}
