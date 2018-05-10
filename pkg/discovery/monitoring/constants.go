package monitoring

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	//EntityType
	ApplicationType        = int32(1)
	VirtualApplicationType = int32(2)
	VirtualMachineType     = int32(3)

	//CommodityType
	TPS     = "tps"
	Latency = "latency"

	//Labels
	IP       = "ip"
	Port     = "port"
	Name     = "name"
	Category = "category"

	//MetricType
	USED     = "used"
	CAPACITY = "capacity"

	//Capacity
	TPSCap     = 20.0
	LatencyCap = 500.0 //ms

	// The default namespace of entity property
	DefaultPropertyNamespace string = "DEFAULT"

	// The attribute used for stitching with other probes (e.g., prometurbo) with app and vapp
	stitchingAttr string = "IP"

	AppPrefix = "app-"
)

var EntityTypeMap = map[int32]proto.EntityDTO_EntityType{
	ApplicationType: proto.EntityDTO_APPLICATION,
}

var CommodityTypeMap = map[string]proto.CommodityDTO_CommodityType{
	TPS:     proto.CommodityDTO_TRANSACTION,
	Latency: proto.CommodityDTO_RESPONSE_TIME,
}

var CommodityCapMap = map[proto.CommodityDTO_CommodityType]float64{
	proto.CommodityDTO_TRANSACTION:     TPSCap,
	proto.CommodityDTO_RESPONSE_TIME: LatencyCap,
}
