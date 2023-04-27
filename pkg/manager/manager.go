package manager

import (
	"crypto/tls"
	"github.com/wangxn2015/onos-lib-go/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"context"
	"fmt"
	"github.com/wangxn2015/onos-lib-go/pkg/certs"
	"sort"
	"strings"
	"time"

	prototypes "github.com/gogo/protobuf/types"
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

	m.HandleRequest()
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

func (m *Manager) HandleRequest() error {
	var types []string
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

	for key := range attr {
		types = append(types, key)
	}
	sort.Strings(types)

	header := fmt.Sprintf("%-10s %20s %20s %15s", nodeIDHeader, cellObjIDHeader, cellGlobalIDHeader, timeHeader)
	//header := fmt.Sprintf("%-10s %20s %20s", "Node ID", "Cell Object ID", "Time")

	for _, key := range types {
		tmpHeader := header
		header = fmt.Sprintf(fmt.Sprintf("%%s %%%ds", len(key)+3), tmpHeader, key)
		//header = fmt.Sprintf("%s %25s", tmpHeader, key)
	}

	//if !noHeaders {
	//	_, _ = fmt.Fprintln(writer, header)
	//}

	keys := make([]string, 0, len(results))
	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, keyID := range keys {
		metrics := results[keyID]
		// sort 2nd map with timestamp
		timeKeySlice := make([]uint64, 0, len(metrics))
		for timeStampKey := range metrics {
			timeKeySlice = append(timeKeySlice, timeStampKey)
		}

		sort.Slice(timeKeySlice, func(i, j int) bool { return timeKeySlice[i] < timeKeySlice[j] })

		for _, timeStamp := range timeKeySlice {
			timeObj := time.Unix(0, int64(timeStamp))
			tsFormat := fmt.Sprintf("%02d:%02d:%02d.%d", timeObj.Hour(), timeObj.Minute(), timeObj.Second(), timeObj.Nanosecond()/1000000)

			ids := strings.Split(keyID, ":")
			e2ID, nodeID, cellID, cellGlobalID := ids[0], ids[1], ids[2], ids[3]
			resultLine := fmt.Sprintf("%-10s %20s %20s %15s", fmt.Sprintf("%s:%s", e2ID, nodeID), cellID, cellGlobalID, tsFormat)
			//resultLine := fmt.Sprintf("%-10s %20s %20s", nodeID, fmt.Sprintf("%x", cellNum), tsFormat)
			for _, typeValue := range types {
				tmpResultLine := resultLine
				var tmpValue string
				if _, ok := metrics[timeStamp][typeValue]; !ok {
					tmpValue = "N/A"
				} else {
					tmpValue = metrics[timeStamp][typeValue]
				}
				resultLine = fmt.Sprintf(fmt.Sprintf("%%s %%%ds", len(typeValue)+3), tmpResultLine, tmpValue)
			}
			//_, _ = fmt.Fprintln(writer, resultLine)
			log.Info(resultLine)
		}
		//_ = writer.Flush()
	}
	return nil
}
