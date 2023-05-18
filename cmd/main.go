//本示例xapp-demo用于演示如何制作一个xapp，用以访问KPM xapp北向接口提供的数据

package main

import (
	"flag"
	"github.com/wangxn2015/onos-lib-go/pkg/certs"
	"github.com/wangxn2015/onos-lib-go/pkg/logging"
	"github.com/wangxn2015/xapp-demo/pkg/manager"
)

const (
	defaultAddress = "192.168.127.113:6789" //!![此处需要配置] 将要连接的， KPM xapp的北向接口
)
const (
	addressKey = "service-address"

	tlsCertPathKey = "tls.certPath"
	tlsKeyPathKey  = "tls.keyPath"
	noTLSKey       = "no-tls"
	authHeaderKey  = "auth-header"
)

var configName string

var configOptions = []string{
	addressKey,
	tlsCertPathKey,
	tlsKeyPathKey,
	noTLSKey,
	authHeaderKey,
}

var log = logging.GetLogger()

func main() {
	logging.SetLevel(logging.DebugLevel)
	caPath := flag.String("caPath", "certs/tls.cacrt", "path to CA certificate")
	keyPath := flag.String("keyPath", "certs/tls.key", "path to client private key")
	certPath := flag.String("certPath", "certs/tls.crt", "path to client certificate")
	kpimonEndpoint := flag.String("kpimonEndpoint", defaultAddress, "kpimon service endpoint")
	noTLSFlag := flag.Bool(noTLSKey, false, "no TLS flag -- true or false")

	ready := make(chan bool)

	flag.Parse()

	// 用于检测证书，无实际功能
	_, err := certs.HandleCertPaths(*caPath, *keyPath, *certPath, true)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Starting xApp-demo")
	cfg := manager.Config{
		CAPath:         *caPath,
		KeyPath:        *keyPath,
		CertPath:       *certPath,
		KpimonEndpoint: *kpimonEndpoint,
		NoTLSFlag:      *noTLSFlag,
	}

	mgr := manager.NewManager(cfg)
	mgr.Run()
	<-ready
}
