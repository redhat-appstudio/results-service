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
	tektonClient := pipelineclientset.New(s.Client.RESTClient())
	trClient := tektonClient.TektonV1().TaskRuns(message.Namespace)
	tr, err := trClient.Get(context.Background(), message.TaskRunName, metav1.GetOptions{})

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
			_, err = trClient.UpdateStatus(context.Background(), tr, metav1.UpdateOptions{})
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
	_, err = trClient.UpdateStatus(context.Background(), tr, metav1.UpdateOptions{})
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	writer.WriteHeader(204)
}
