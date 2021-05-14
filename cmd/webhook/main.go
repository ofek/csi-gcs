package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ofek/csi-gcs/pkg/driver"
	"github.com/ofek/csi-gcs/pkg/util"
	"github.com/ofek/csi-gcs/pkg/webhook"
	"k8s.io/klog"
)

var (
	driverNameFlag = flag.String("driver-name", driver.CSIDriverName, "CSI driver name")
	versionFlag    = flag.Bool("version", false, "Print the version and exit")
	listenAddrFlag = flag.String("listen", ":9443", "Listen address for the HTTP webhook server")
	tlsCertFlag    = flag.String("tls-crt", "/tls/tls.crt", "TLS certificate for the HTTP webhook server")
	tlsKeyFlag     = flag.String("tls-key", "/tls/tls.key", "TLS key for the HTTP webhook server")
)

func main() {
	_ = flag.Set("alsologtostderr", "true")
	klog.InitFlags(nil)
	util.SetEnvVarFlags()
	flag.Parse()

	if *versionFlag {
		versionJSON, err := driver.GetVersionJSON()
		if err != nil {
			klog.Exit(err.Error())
		}
		fmt.Println(versionJSON)
		return
	}

	wbk, err := webhook.NewServer(*driverNameFlag)
	if err != nil {
		klog.Exitf("Unable to create webhook server: %+v", err)
	}
	srv := &http.Server{
		Addr:    *listenAddrFlag,
		Handler: wbk,
		// mutating webhooks should answer within milliseconds, those timeouts should be more than enough.
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		klog.V(3).Info("Server Starting")
		if err := srv.ListenAndServeTLS(*tlsCertFlag, *tlsKeyFlag); err != nil && err != http.ErrServerClosed {
			klog.Exitf("Server stopped unexpectedly: %+v", err)
		}
	}()

	<-done
	klog.V(3).Info("Server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		klog.Exitf("Server stopped with error: %+v", err)
	}
	klog.V(3).Info("Server stopped")
}
