// https-server.go
package main

import (
	"crypto/tls"
	"flag"
	"github.com/redhat-appstudio/results-service/pkg/resultsservice"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"log"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

func main() {

	klog.InitFlags(flag.CommandLine)

	flag.Parse()
	opts := zap.Options{
		TimeEncoder: zapcore.RFC3339TimeEncoder,
		ZapOpts:     []zap2.Option{zap2.WithCaller(true)},
	}
	logger := zap.New(zap.UseFlagOptions(&opts))

	mainLog := logger.WithName("main")
	klog.SetLogger(mainLog)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// load tls certificates
	serverTLSCert, err := tls.LoadX509KeyPair(CertFilePath, KeyFilePath)
	if err != nil {
		log.Fatalf("Error loading certificate and key file: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/store-results", &resultsservice.ResultsService{Logger: &mainLog, Client: clientset})

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
		MinVersion:   tls.VersionTLS12,
	}
	logger.Info("starting HTTP server")

	server := http.Server{
		Addr:              ":8443",
		Handler:           mux,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: time.Second * 3,
	}
	defer server.Close()
	log.Fatal(server.ListenAndServeTLS("", ""))
}
