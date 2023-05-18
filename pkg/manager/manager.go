package manager

import (
	"context"
	"crypto/tls"
	"fmt"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/wangxn2015/onos-lib-go/pkg/certs"
	"github.com/wangxn2015/onos-lib-go/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	//"github.com/prometheus/common/log"

	kpimonapi "github.com/wangxn2015/xapp-demo/api/onos.kpimon"
)

const (
	nodeIDHeader       = "Node ID"
	cellObjIDHeader    = "Cell Object ID"
	cellGlobalIDHeader = "Cell Global ID"
	timeHeader         = "Time"
)

var log = logging.GetLogger()

// Config is a manager configuration
type Config struct {
	CAPath         string
	KeyPath        string
	CertPath       string
	KpimonEndpoint string
	NoTLSFlag      bool
	//GRPCPort    int
	//RicActionID int32
	//ConfigPath  string
}

// NewManager generates the new xAPP manager
func NewManager(config Config) *Manager {

	manager := &Manager{

		config: config,
	}
	return manager
}

// Manager is an abstract struct for manager
type Manager struct {
	config Config
	conn   *grpc.ClientConn
}

// Run runs KPIMON manager
func (m *Manager) Run() {
	err := m.start()
	if err != nil {
		log.Errorf("Error when starting xapp: %v", err)
	}
}

// Close closes manager
func (m *Manager) Close() {
	log.Info("closing Manager")
}

func (m *Manager) start() error {
	err := m.startClient()
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (m *Manager) startClient() error {
	conn, err := m.GetConnection()
	if err != nil {
		log.Warn(err)
		return err
	}
	m.conn = conn

	m.HandleRequestRepeated()

	log.Info("closing conn")
	conn.Close()
	return nil
}

// GetConnection returns a gRPC client connection to the onos service
func (m *Manager) GetConnection() (*grpc.ClientConn, error) {
	address := m.config.KpimonEndpoint
	certPath := m.config.CertPath
	keyPath := m.config.KeyPath
	var opts []grpc.DialOption

	if m.config.NoTLSFlag {
		opts = []grpc.DialOption{
			grpc.WithInsecure(),
		}
	} else { //enter here since set  NoTLS to false
		log.Info("load TLS config...")
		if certPath != "" && keyPath != "" {
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return nil, err
			}
			opts = []grpc.DialOption{
				grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
					Certificates:       []tls.Certificate{cert},
					InsecureSkipVerify: true,
				})),
			}
		} else {
			// Load default Certificates
			cert, err := tls.X509KeyPair([]byte(certs.DefaultClientCrt), []byte(certs.DefaultClientKey))
			if err != nil {
				return nil, err
			}
			opts = []grpc.DialOption{
				grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
					Certificates:       []tls.Certificate{cert},
					InsecureSkipVerify: true,
				})),
			}
		}
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Info(err)
		return nil, err
	}
	return conn, nil
}

func (m *Manager) HandleRequestRepeated() error {
	//var types []string
	var results map[string]map[uint64]map[string]string

	request := kpimonapi.GetRequest{}
	client := kpimonapi.NewKpimonClient(m.conn)

	respWatchMeasurement, err := client.WatchMeasurements(context.Background(), &request)
	if err != nil {
		return err
	}

	for {
		respGetMeasurement, err := respWatchMeasurement.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		results = make(map[string]map[uint64]map[string]string)

		attr := make(map[string]string)
		for key, measItems := range respGetMeasurement.GetMeasurements() {
			//在目前的配置中，measItems.MeasurementItems测量重复共2次，见打印中的ii=0,1
			for ii, measItem := range measItems.MeasurementItems {
				fmt.Printf("ii= %d\n", ii)
				for _, measRecord := range measItem.MeasurementRecords {
					timeStamp := measRecord.Timestamp
					measName := measRecord.MeasurementName
					measValue := measRecord.MeasurementValue

					if _, ok := attr[measName]; !ok {
						attr[measName] = measName
					}

					if _, ok1 := results[key]; !ok1 {
						results[key] = make(map[uint64]map[string]string)
					}
					if _, ok2 := results[key][timeStamp]; !ok2 {
						results[key][timeStamp] = make(map[string]string)
					}

					var value interface{}
					switch {
					case prototypes.Is(measValue, &kpimonapi.IntegerValue{}):
						v := kpimonapi.IntegerValue{}
						err := prototypes.UnmarshalAny(measValue, &v)
						if err != nil {
							log.Warn(err)
						}
						value = v.GetValue()
						//fmt.Printf("%s\t: %d\t"+"time: %v\t %v\n", measName, value, timeStamp, time.Unix(0, int64(timeStamp)))
						fmt.Printf("%s\t: %d\t"+"time: %v\t \n", measName, value, timeStamp)

					case prototypes.Is(measValue, &kpimonapi.RealValue{}):
						v := kpimonapi.RealValue{}
						err := prototypes.UnmarshalAny(measValue, &v)
						if err != nil {
							log.Warn(err)
						}
						value = v.GetValue()
						//fmt.Printf("%s\t: %f\t"+"time: %v\t %v\n", measName, value, timeStamp, time.Unix(0, int64(timeStamp)))
						fmt.Printf("%s\t: %f\t"+"time: %v\t \n", measName, value, timeStamp)
					case prototypes.Is(measValue, &kpimonapi.NoValue{}):
						v := kpimonapi.NoValue{}
						err := prototypes.UnmarshalAny(measValue, &v)
						if err != nil {
							log.Warn(err)
						}
						value = v.GetValue()
						//fmt.Printf("%s\t: N/A %v\t"+"time: %v\t %v\n", measName, value, timeStamp, time.Unix(0, int64(timeStamp)))
						fmt.Printf("%s\t: N/A %v\t"+"time: %v\t \n", measName, value, timeStamp)

					}

					results[key][timeStamp][measName] = fmt.Sprintf("%v", value)
				}
			}
		}
		fmt.Printf("-------\n")
		//或者可以使用results结构，一次性打印，如下
		log.Info("Data received is: ", results)
		//-----------------------------------

	}
	return nil
}

func (m *Manager) HandleRequestForOneTimeRequest() error {
	//var types []string
	results := make(map[string]map[uint64]map[string]string)

	request := kpimonapi.GetRequest{}
	client := kpimonapi.NewKpimonClient(m.conn)
	log.Info("HandleRequest")
	respGetMeasurement, err := client.ListMeasurements(context.Background(), &request)
	if err != nil {
		log.Info("err: ", err)
		return err
	}
	fmt.Printf("response: ", respGetMeasurement)
	log.Info("response: ", respGetMeasurement)

	attr := make(map[string]string)
	for key, measItems := range respGetMeasurement.GetMeasurements() {
		for _, measItem := range measItems.MeasurementItems {
			for _, measRecord := range measItem.MeasurementRecords {
				timeStamp := measRecord.Timestamp
				measName := measRecord.MeasurementName
				measValue := measRecord.MeasurementValue

				if _, ok := attr[measName]; !ok {
					attr[measName] = measName
				}

				if _, ok1 := results[key]; !ok1 {
					results[key] = make(map[uint64]map[string]string)
				}
				if _, ok2 := results[key][timeStamp]; !ok2 {
					results[key][timeStamp] = make(map[string]string)
				}

				var value interface{}

				switch {
				case prototypes.Is(measValue, &kpimonapi.IntegerValue{}):
					v := kpimonapi.IntegerValue{}
					err := prototypes.UnmarshalAny(measValue, &v)
					if err != nil {
						log.Warn(err)
					}
					value = v.GetValue()

				case prototypes.Is(measValue, &kpimonapi.RealValue{}):
					v := kpimonapi.RealValue{}
					err := prototypes.UnmarshalAny(measValue, &v)
					if err != nil {
						log.Warn(err)
					}
					value = v.GetValue()

				case prototypes.Is(measValue, &kpimonapi.NoValue{}):
					v := kpimonapi.NoValue{}
					err := prototypes.UnmarshalAny(measValue, &v)
					if err != nil {
						log.Warn(err)
					}
					value = v.GetValue()

				}
				results[key][timeStamp][measName] = fmt.Sprintf("%v", value)
			}
		}
	}

	return nil
}
