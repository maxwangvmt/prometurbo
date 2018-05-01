package registration

import (
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	TargetIdField string = "targetIdentifier"
	ProbeCategory string = "Cloud Native"
	TargetType    string = "Prometheus"
)

// Registration Client for the Prometheus probe
// Implements the TurboRegistrationClient interface
type PrometheusRegistrationClient struct {
}

func (myProbe *PrometheusRegistrationClient) GetSupplyChainDefinition() []*proto.TemplateDTO {
	glog.Infoln("Building a supply chain ..........")

	// 2. Build supply chain.
	supplyChainFactory := &SupplyChainFactory{}
	templateDtos, err := supplyChainFactory.CreateSupplyChain()
	if err != nil {
		glog.Infoln("Error creating Supply chain for the Prometheus probe")
		return nil
	}
	glog.Infoln("Supply chain for the Prometheus probe is created.")
	return templateDtos
}

func (registrationClient *PrometheusRegistrationClient) GetIdentifyingFields() string {
	return TargetIdField
}

func (myProbe *PrometheusRegistrationClient) GetAccountDefinition() []*proto.AccountDefEntry {

	targetIDAcctDefEntry := builder.NewAccountDefEntryBuilder(TargetIdField, "URL",
		"URL of the Prometheus target", ".*", true, false).Create()

	return []*proto.AccountDefEntry{
		targetIDAcctDefEntry,
	}
}
