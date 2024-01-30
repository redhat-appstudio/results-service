package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/redhat-appstudio/results-service/pkg/resultsservice"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {

	klog.InitFlags(flag.CommandLine)
	var resultsServiceUri string
	var resultsDirectory string
	var taskRunName string
	var namespace string

	flag.StringVar(&resultsServiceUri, "results-service", "", "The URI or the results service")
	flag.StringVar(&resultsDirectory, "results-directory", "/large-results", "The location of the alternate results directory")
	flag.StringVar(&taskRunName, "task-run-name", "", "The task run name, should be the value of $(context.taskRun.name)")
	flag.StringVar(&namespace, "task-run-namespace", "", "The task run name, should be the value of $(context.taskRun.namespace)")
	opts := zap.Options{
		TimeEncoder: zapcore.RFC3339TimeEncoder,
		ZapOpts:     []zap2.Option{zap2.WithCaller(true)},
	}
	opts.BindFlags(flag.CommandLine)
	klog.InitFlags(flag.CommandLine)
	flag.Parse()
	if resultsServiceUri == "" || resultsDirectory == "" {
		println("Must specify both results-service and results-directory params")
		os.Exit(1)
	}

	logger := zap.New(zap.UseFlagOptions(&opts))

	mainLog := logger.WithName("main")
	klog.SetLogger(mainLog)

	files, err := os.ReadDir(resultsDirectory)
	if err != nil {
		panic(err)
	}
	data := resultsservice.ResultsMessage{
		Results:     map[string]string{},
		TaskRunName: taskRunName,
		Namespace:   namespace,
	}
	for _, resultsFile := range files {
		fileContents, err := os.ReadFile(resultsDirectory + "/" + resultsFile.Name())
		if err != nil {
			logger.Error(err, "Failed to read results file", "file", resultsFile.Name())
			panic(err)
		}
		data.Results[resultsFile.Name()] = string(fileContents)
	}
	marshalled, err := json.Marshal(&data)

	response, err := http.Post(resultsServiceUri, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		logger.Error(err, "HTTP request failed")
		panic(err)
	}
	if response.StatusCode != 204 {
		err := fmt.Errorf("expecting response 204, got %d", response.StatusCode)
		logger.Error(err, "HTTP request failed")
		panic(err)
	}
}
