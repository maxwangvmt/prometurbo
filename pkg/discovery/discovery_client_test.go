package discovery

import (
	"reflect"
	"testing"

	"github.com/turbonomic/prometurbo/pkg/discovery/monitoring"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"fmt"
	"math"
)

var (
	namespace = "DEFAULT"
	ipAttr = "IP"
	appPrefix = "APPLICATION-"
	appType = proto.EntityDTO_APPLICATION
	useTopoExt = true
	scope = "k8s-cluster-foo"
	targetAddr = "target-foo"

	replacementMetaData = builder.NewReplacementEntityMetaDataBuilder().
				Matching(ipAttr).
				MatchingExternal(&proto.ServerEntityPropDef{
							Entity:     &appType,
							Attribute:  &ipAttr,
							UseTopoExt: &useTopoExt,
						}).
				PatchSellingWithProperty(proto.CommodityDTO_TRANSACTION, []string{"used", "capacity"}).
				PatchSellingWithProperty(proto.CommodityDTO_RESPONSE_TIME, []string{"used", "capacity"}).
				Build()


	metrics = []*monitoring.EntityMetric{
		newMetric("1.2.3.4", 13.4, 66.7, monitoring.ApplicationType),
		newMetric("5.6.7.8", 0, 0, monitoring.ApplicationType),
		newMetric("15.16.17.18", monitoring.TPSCap, monitoring.LatencyCap, monitoring.ApplicationType),
		newMetric("115.116.117.118", monitoring.TPSCap+10, monitoring.LatencyCap+10, monitoring.ApplicationType),
	}
)

func TestP8sDiscoveryClient_GetAccountValues(t *testing.T) {
	d := NewDiscoveryClient(targetAddr, scope, []monitoring.MetricExporter{})

	for _, f := range d.GetAccountValues().GetTargetInstance().InputFields {
		if f.Name == "targetIdentifier" && f.Value == targetAddr {
			return
		}
	}

	t.Errorf("AccountValues does not contain targetIdentifier with value %s", targetAddr)
}

func TestP8sDiscoveryClient_Discover(t *testing.T) {
	exporter := &mockExporter{
		metrics: metrics,
	}

	d := NewDiscoveryClient(targetAddr, scope, []monitoring.MetricExporter{exporter})

	testDiscoverySuccedded(d, metrics)
}

func TestP8sDiscoveryClient_Discover_Two_Exporters(t *testing.T) {
	exporter1 := &mockExporter{
		metrics: metrics[0:2],
	}

	exporter2 := &mockExporter{
		metrics: metrics[2:],
	}

	d := NewDiscoveryClient(targetAddr, scope, []monitoring.MetricExporter{exporter1, exporter2})

	testDiscoverySuccedded(d, metrics)
}

func TestP8sDiscoveryClient_Discover_Two_Exporters_One_Failed(t *testing.T) {
	exporter1 := &mockExporter{
		metrics: metrics[0:2],
	}

	exporter2 := &mockExporter{
		err: fmt.Errorf("Query failed with the mocked exporter"),
	}

	d := NewDiscoveryClient(targetAddr, scope, []monitoring.MetricExporter{exporter1, exporter2})

	testDiscoverySuccedded(d, metrics[0:2])
}

func TestP8sDiscoveryClient_Discover_Query_Failed(t *testing.T) {
	exporter1 := &mockExporter{
		err: fmt.Errorf("Query failed with the mocked exporter"),
	}

	exporter2 := &mockExporter{
		err: fmt.Errorf("Query failed with the mocked exporter"),
	}

	d := NewDiscoveryClient(targetAddr, scope, []monitoring.MetricExporter{exporter1, exporter2})

	res, err := d.Discover([]*proto.AccountValue{})

	if err != nil {
		t.Errorf("P8sDiscoveryClient.Discover() error = %v", err)
		return
	}

	// Expect to see error in the response
	if len(res.ErrorDTO) != 1 || *res.ErrorDTO[0].Severity != proto.ErrorDTO_CRITICAL {
		t.Errorf("Expected one error DTO with serverity CRITICAL but got %v", res.ErrorDTO)
	}
}

type mockExporter struct {
	metrics []*monitoring.EntityMetric
	err error
}

func (m *mockExporter) Query() ([]*monitoring.EntityMetric, error) {
	return m.metrics, m.err
}

func newMetric(ip string, tpsUsed, latUsed float64, entityType int32) *monitoring.EntityMetric {
	m := map[string]float64{
		monitoring.TPS: tpsUsed,
		monitoring.Latency: latUsed,
	}

	return &monitoring.EntityMetric{
		UID: ip,
		Type: entityType,
		Metrics: m,
	}
}

func checkAppResult(metric *monitoring.EntityMetric, entity *proto.EntityDTO) error {
	ip := metric.UID
	tpsUsed := metric.Metrics[monitoring.TPS]
	latUsed := metric.Metrics[monitoring.Latency]

	commodities := []*proto.CommodityDTO{
		newTrasactionCommodity(tpsUsed, ip),
		newResponseTimeCommodity(latUsed, ip),
	}

	entityProperty := &proto.EntityDTO_EntityProperty{
		Namespace: &namespace,
		Name:      &ipAttr,
		Value:     &ip,
	}

	dto, err := builder.NewEntityDTOBuilder(proto.EntityDTO_APPLICATION, appPrefix+scope+"/"+ip).
				DisplayName(appPrefix+scope+"/"+ip).
				SellsCommodities(commodities).
				WithProperty(entityProperty).
				ReplacedBy(replacementMetaData).
				Create()

	if err != nil {
		return err
	}

	if !reflect.DeepEqual(dto, entity) {
		return fmt.Errorf("The entity %v doesn't match the metric %v", entity, metric)
	}

	return nil
}

func newTrasactionCommodity(used float64, key string) *proto.CommodityDTO {
	capacity := math.Max(used, monitoring.TPSCap)
	comm, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_TRANSACTION).
		Used(used).Capacity(capacity).Key(key).Create()
	return comm
}

func newResponseTimeCommodity(used float64, key string) *proto.CommodityDTO {
	capacity := math.Max(used, monitoring.LatencyCap)
	comm, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_RESPONSE_TIME).
		Used(used).Capacity(capacity).Key(key).Create()
	return comm
}

func testDiscoverySuccedded(d *P8sDiscoveryClient, metrics []*monitoring.EntityMetric) error {
	res, err := d.Discover([]*proto.AccountValue{})

	if err != nil {
		return err
	}

	// Check results
	if len(res.ErrorDTO) != 0 {
		return fmt.Errorf("ErrorDTO is not empty: %v", res.ErrorDTO[0])
	}

	if len(res.EntityDTO) != len(metrics) {
		return fmt.Errorf("Expected %d entities but got %d entities", len(metrics), len(res.EntityDTO))
	}

	for i := range metrics {
		if err := checkAppResult(metrics[i], res.EntityDTO[i]); err != nil {
			return fmt.Errorf("P8sDiscoveryClient.Discover() error = %v", err)
		}
	}

	return nil
}
