package conf

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
	"io/ioutil"
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
	metaConfig, err := readConfig(targetConfigFilePath)

	if err != nil {
		return nil, err
	}

	if metaConfig.MetricExporterEndpoint == "" {
		metaConfig.MetricExporterEndpoint = defaultEndpoint
	}

	return metaConfig, nil
}

// Get the config from file.
func readConfig(path string) (*PrometheusConf, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("File error: %v\n", err)
		return nil, err
	}
	glog.Infoln(string(file))

	var config PrometheusConf
	err = json.Unmarshal(file, &config)

	if err != nil {
		glog.Errorf("Unmarshall error :%v\n", err)
		return nil, err
	}
	glog.Infof("Results: %+v\n", config)

	return &config, nil
}
