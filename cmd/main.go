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
	configDir  = ".onos"
	addressKey = "service-address"

	tlsCertPathKey = "tls.certPath"
	tlsKeyPathKey  = "tls.keyPath"
	noTLSKey       = "no-tls"
	authHeaderKey  = "auth-header"

	addressFlag     = "service-address"
	tlsCertPathFlag = "tls-cert-path"
	tlsKeyPathFlag  = "tls-key-path"
	noTLSFlag       = "no-tls"
	// AuthHeaderFlag - the flag name
	AuthHeaderFlag = "auth-header"

	// Authorization the header keyword
	Authorization = "authorization"
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
	//ricActionID := flag.Int("ricActionID", 10, "RIC Action ID in E2 message")
	//configPath := flag.String("configPath", "/etc/onos/config/config.json", "path to config.json file")
	//grpcPort := flag.Int("grpcPort", 5150, "grpc Port number")

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
		//GRPCPort:    *grpcPort,
		//RicActionID: int32(*ricActionID),
		//ConfigPath:  *configPath,
	}

	mgr := manager.NewManager(cfg)
	mgr.Run()
	<-ready
}
