package monitoring

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"math"
)

type IEntityBuilder interface {
	Build(metric *EntityMetric) ([]*proto.EntityDTO, error)
}

type appEntityBuilder struct {
	scope string
}

func NewAppEntityBuilder(scope string) *appEntityBuilder {
	return &appEntityBuilder{scope}
}

func (b *appEntityBuilder) Build(metric *EntityMetric) ([]*proto.EntityDTO, error) {
	ipAttr := "IP" //supplychain.SUPPLY_CHAIN_CONSTANT_IP_ADDRESS
	ns := "DEFAULT"
	useTopoExt := true
	ip := metric.UID

	entityType, ok := EntityTypeMap[metric.Type]
	if !ok {
		err := fmt.Errorf("Unsupported entity type %v", metric.Type)
		glog.Errorf(err.Error())
		return nil, err
	}

	// Construct the commodity
	replaceBuilder := builder.NewReplacementEntityMetaDataBuilder()
	replaceBuilder.Matching(ipAttr)
	replaceBuilder.MatchingExternal(&proto.ServerEntityPropDef{
		Entity:     &entityType,
		Attribute:  &ipAttr,
		UseTopoExt: &useTopoExt,
	})

	commodities := []*proto.CommodityDTO{}
	commMetrics := metric.Metrics
	capacity := 100.0
	for key, value := range commMetrics {
		var commType proto.CommodityDTO_CommodityType
		commType, ok := CommodityTypeMap[key]

		if !ok {
			err := fmt.Errorf("Unsupported commodity type %s", key)
			glog.Errorf(err.Error())
			continue
		}

		if commType == proto.CommodityDTO_RESPONSE_TIME {
			capacity = LatencyCap
			value *= 1000 // Convert second to millisecond
			if value >= LatencyCap {
				//value = LatencyCap
				capacity = value + 1
			}
		} else if commType == proto.CommodityDTO_TRANSACTION {
			capacity = TPSCap
			value = math.Min(value, TPSCap)
			if value >= TPSCap {
				//value = TPSCap - 1
				capacity = value + 1
			}
		}

		commodity, err := builder.NewCommodityDTOBuilder(commType).
			Used(value).Capacity(capacity).Key(ip).Create()

		if err != nil {
			glog.Errorf("Error building a commodity: %s", err)
			continue
		}

		commodities = append(commodities, commodity)
		replaceBuilder.PatchSellingWithProperty(commType, []string{USED, CAPACITY})
	}

	if len(commodities) < 2 {
		c, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_RESPONSE_TIME).
			Used(0.0).Capacity(capacity).Key(ip).Create()
		commodities = append(commodities, c)
		// TODO
		replaceBuilder.PatchSellingWithProperty(proto.CommodityDTO_RESPONSE_TIME, []string{USED, CAPACITY})
	}

	dto, err := builder.NewEntityDTOBuilder(entityType, "App/"+ip).
		DisplayName("max-api-" + ip).
		SellsCommodities(commodities).
		WithProperty(&proto.EntityDTO_EntityProperty{
			Namespace: &ns,
			Name:      &ipAttr,
			Value:     &ip,
		}).
		ReplacedBy(replaceBuilder.Build()).
		Create()

	if err != nil {
		glog.Errorf("Error building EntityDTO from metric %v: %s", metric, err)
		return nil, err
	}

	dtos := []*proto.EntityDTO{dto}

	return dtos, nil
}
