package discovery

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/prometurbo/pkg/discovery/monitoring"
	"github.com/turbonomic/prometurbo/pkg/registration"
)

// Discovery Client for the Prometheus Probe
// Implements the TurboDiscoveryClient interface
type PrometheusDiscoveryClient struct {
	targetAddr       string
	metricExporters  []monitoring.IMetricExporter
	appEntityBuilder monitoring.IEntityBuilder
}

func NewDiscoveryClient(targetAddr string, appEntityBuilder monitoring.IEntityBuilder, metricExporters []monitoring.IMetricExporter) *PrometheusDiscoveryClient {
	return &PrometheusDiscoveryClient{
		targetAddr:       targetAddr,
		metricExporters:  metricExporters,
		appEntityBuilder: appEntityBuilder,
	}
}

// Get the Account Values to create VMTTarget in the turbo server corresponding to this client
func (discClient *PrometheusDiscoveryClient) GetAccountValues() *probe.TurboTargetInfo {
	targetId := registration.TargetIdField
	targetIdVal := &proto.AccountValue{
		Key:         &targetId,
		StringValue: &discClient.targetAddr,
	}

	accountValues := []*proto.AccountValue{
		targetIdVal,
	}

	targetInfo := probe.NewTurboTargetInfoBuilder(registration.ProbeCategory, registration.TargetType,
		registration.TargetIdField, accountValues).Create()

	return targetInfo
}

// Validate the Target
func (discClient *PrometheusDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	glog.V(2).Infof("BEGIN Validation for PrometheusDiscoveryClient %s\n", accountValues)

	validationResponse := &proto.ValidationResponse{}

	glog.V(2).Infof("Validation response %s\n", validationResponse)
	return validationResponse, nil
}

// Discover the Target Topology
func (d *PrometheusDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	glog.V(2).Infof("========= Discovering Prometheus ============= %s\n", accountValues)
	var entities []*proto.EntityDTO

	for _, metricExporter := range d.metricExporters {
		metrics, err := metricExporter.Query()
		if err != nil {
			glog.Errorf("Error while querying metrics exporter: %s\n", err)
			// If there is error during discovery, return an ErrorDTO.
			severity := proto.ErrorDTO_CRITICAL
			description := fmt.Sprintf("%v", err)
			errorDTO := &proto.ErrorDTO{
				Severity:    &severity,
				Description: &description,
			}
			discoveryResponse := &proto.DiscoveryResponse{
				ErrorDTO: []*proto.ErrorDTO{errorDTO},
			}
			return discoveryResponse, nil
		}

		for _, metric := range metrics {
			//if metric.Type != proto.EntityDTO_APPLICATION {
			//	glog.Errorf("Only Application type is supported %v: %s", metric.Type)
			//	continue
			//}
			dtos, err := d.appEntityBuilder.Build(metric)
			if err != nil {
				glog.Errorf("Error building entity from metric %v: %s", metric, err)
				continue
			}
			entities = append(entities, dtos...)
			if len(dtos) > 0 {
				fmt.Printf("Discovered DTO: %++v\n", dtos[0])
			}
		}
	}

	discoveryResponse := &proto.DiscoveryResponse{
		EntityDTO: entities,
	}
	fmt.Printf("Prometheus discovery response %s\n", discoveryResponse)

	return discoveryResponse, nil
}
