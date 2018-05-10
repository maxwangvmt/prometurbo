package discovery

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/prometurbo/pkg/discovery/dtofactory"
	"github.com/turbonomic/prometurbo/pkg/registration"
	"github.com/turbonomic/prometurbo/pkg/discovery/exporter"
)

// Implements the TurboDiscoveryClient interface
type P8sDiscoveryClient struct {
	targetAddr       string
	scope string
	metricExporters  []exporter.MetricExporter
}

func NewDiscoveryClient(targetAddr, scope string, metricExporters []exporter.MetricExporter) *P8sDiscoveryClient {
	return &P8sDiscoveryClient{
		targetAddr:       targetAddr,
		scope: scope,
		metricExporters:  metricExporters,
	}
}

// Get the Account Values to create VMTTarget in the turbo server corresponding to this client
func (d *P8sDiscoveryClient) GetAccountValues() *probe.TurboTargetInfo {
	targetId := registration.TargetIdField
	targetIdVal := &proto.AccountValue{
		Key:         &targetId,
		StringValue: &d.targetAddr,
	}

	scope := registration.Scope
	scopeVal := &proto.AccountValue{
		Key: &scope,
		StringValue: &d.scope,
	}

	accountValues := []*proto.AccountValue{
		targetIdVal,
		scopeVal,
	}

	targetInfo := probe.NewTurboTargetInfoBuilder(registration.ProbeCategory, registration.TargetType,
		registration.TargetIdField, accountValues).Create()

	return targetInfo
}

// Validate the Target
func (d *P8sDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	// TODO: Add logic for validation
	validationResponse := &proto.ValidationResponse{}
	return validationResponse, nil
}

// Discover the Target Topology
func (d *P8sDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	glog.V(2).Infof("Discovering the target %s", accountValues)
	var entities []*proto.EntityDTO
	allExportersFailed := true

	for _, metricExporter := range d.metricExporters {
		dtos, err := d.buildEntities(metricExporter)
		if err != nil {
			glog.Errorf("Error while querying metrics exporter %v: %v", metricExporter, err)
			continue
		}
		allExportersFailed = false
		entities = append(entities, dtos...)

		glog.V(4).Infof("Entities built from exporter %v: %v", metricExporter, dtos)

		//metrics, err := metricExporter.Query()
		//if err != nil {
		//	glog.Errorf("Error while querying metrics exporter: %s", err)
		//	// If there is error during discovery, return an ErrorDTO.
		//	severity := proto.ErrorDTO_CRITICAL
		//	description := fmt.Sprintf("%v", err)
		//	errorDTO := &proto.ErrorDTO{
		//		Severity:    &severity,
		//		Description: &description,
		//	}
		//	discoveryResponse := &proto.DiscoveryResponse{
		//		ErrorDTO: []*proto.ErrorDTO{errorDTO},
		//	}
		//	return discoveryResponse, nil
		//}
		//
		//for _, metric := range metrics {
		//	//if metric.Type != proto.EntityDTO_APPLICATION {
		//	//	glog.Errorf("Only Application type is supported %v: %s", metric.Type)
		//	//	continue
		//	//}
		//	dtos, err := d.entityBuilder.Build(metric)
		//	if err != nil {
		//		glog.Errorf("Error building entity from metric %v: %s", metric, err)
		//		continue
		//	}
		//	entities = append(entities, dtos...)
		//	//if len(dtos) > 0 {
		//	//	glog.V(4).Infof("Discovered DTO: %++v", dtos[0])
		//	//}
		//}
	}

	// The discovery fails if all queries to exporters fail
	if allExportersFailed {
		return d.failDiscovery(), nil
	}

	discoveryResponse := &proto.DiscoveryResponse{
		EntityDTO: entities,
	}
	//glog.V(4).Infof("Discovery response %v", discoveryResponse)

	return discoveryResponse, nil
}

func (d *P8sDiscoveryClient) buildEntities(metricExporter exporter.MetricExporter) ([]*proto.EntityDTO, error) {
	var entities []*proto.EntityDTO

	metrics, err := metricExporter.Query()
	if err != nil {
		glog.Errorf("Error while querying metrics exporter: %v", err)
		return nil, err
	}

	for _, metric := range metrics {
		dtos, err := dtofactory.NewEntityBuilder(d.scope, metric).Build()
		if err != nil {
			glog.Errorf("Error building entity from metric %v: %s", metric, err)
			continue
		}
		entities = append(entities, dtos...)
	}

	return entities, nil
}

func (d *P8sDiscoveryClient) failDiscovery() *proto.DiscoveryResponse {
	description := fmt.Sprintf("All exporter queries failed: %v", d.metricExporters)
	glog.Errorf(description)
	// If there is error during discovery, return an ErrorDTO.
	severity := proto.ErrorDTO_CRITICAL
	errorDTO := &proto.ErrorDTO{
		Severity:    &severity,
		Description: &description,
	}
	discoveryResponse := &proto.DiscoveryResponse{
		ErrorDTO: []*proto.ErrorDTO{errorDTO},
	}
	return discoveryResponse
}
