package conf

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
	"io/ioutil"
)

const (
	LocalDebugConfPath = "configs/prometurbo-config"
	DefaultConfPath    = "/etc/prometurbo/turbo.config"
	defaultEndpoint    = "http://localhost:8081/pod/metrics"
)

type PrometurboConf struct {
	Communicator           *service.TurboCommunicationConfig `json:"communicationConfig,omitempty"`
	TargetConf             *PrometurboTargetConf             `json:"prometurboTargetConfig,omitempty"`
	MetricExporterEndpoint string                            `json:"metricExporterEndpoint,omitempty"`
}

type PrometurboTargetConf struct {
	Address string `json:"targetAddress,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

func NewPrometurboConf(serviceConfigFilePath string) (*PrometurboConf, error) {

	glog.Infof("Read configuration from %s", serviceConfigFilePath)
	metaConfig, err := readConfig(serviceConfigFilePath)

	if err != nil {
		return nil, err
	}

	if metaConfig.MetricExporterEndpoint == "" {
		metaConfig.MetricExporterEndpoint = defaultEndpoint
	}

	return metaConfig, nil
}

func readConfig(path string) (*PrometurboConf, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("File error: %v\n", err)
		return nil, err
	}
	glog.Infoln(string(file))

	var config PrometurboConf
	err = json.Unmarshal(file, &config)

	if err != nil {
		glog.Errorf("Unmarshall error :%v\n", err)
		return nil, err
	}
	glog.Infof("Results: %+v\n", config)

	return &config, nil
}
