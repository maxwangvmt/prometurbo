package conf

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
	"io/ioutil"
	"os"
)

const (
	LocalDebugConfPath = "configs/conf.json"
	DefaultConfPath    = "/etc/prometurbo/turbo.config"
	defaultEndpoint    = "http://localhost:8081/pod/metrics"
)

type PrometheusConf struct {
	Communicator           *service.TurboCommunicationConfig `json:"communicationConfig,omitempty"`
	TargetConf             *PrometheusTargetConf             `json:"targetConfig,omitempty"`
	MetricExporterEndpoint string                            `json:"metricExporterEndpoint,omitempty"`
}

type PrometheusTargetConf struct {
	Address string `json:"targetAddress,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

func NewPrometheusConf(targetConfigFilePath string) (*PrometheusConf, error) {

	glog.Infof("Read configuration from %s\n", targetConfigFilePath)
	metaConfig := readConfig(targetConfigFilePath)

	if metaConfig.MetricExporterEndpoint == "" {
		metaConfig.MetricExporterEndpoint = defaultEndpoint
	}

	return metaConfig, nil
}

// Get the config from file.
func readConfig(path string) *PrometheusConf {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		glog.Infof("File error: %v\n", e)
		os.Exit(1)
	}
	glog.Infoln(string(file))

	var config PrometheusConf
	err := json.Unmarshal(file, &config)

	if err != nil {
		glog.Errorf("Unmarshall error :%v\n", err)
	}
	glog.Infof("Results: %+v\n", config)

	return &config
}
