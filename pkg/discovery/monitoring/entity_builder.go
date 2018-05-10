package monitoring

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type entityBuilder struct {
	// TODO: Add the scope to the property for stitching, which needs corresponding change at kubeturbo side
	scope string

	metric *EntityMetric
}

func NewEntityBuilder(scope string, metric *EntityMetric) *entityBuilder {
	return &entityBuilder{
		scope: scope,
		metric: metric,
	}
}

func (b *entityBuilder) Build() ([]*proto.EntityDTO, error) {
	metric := b.metric
	//ipAttr := AppStitchingAttr
	//ns := DefaultPropertyNamespace
	//useTopoExt := true

	entityType, ok := EntityTypeMap[metric.Type]
	if !ok {
		err := fmt.Errorf("Unsupported entity type %v", metric.Type)
		glog.Errorf(err.Error())
		return nil, err
	}

	ip := metric.UID

	// Construct the commodity
	//replaceBuilder := builder.NewReplacementEntityMetaDataBuilder()
	//replaceBuilder.Matching(ipAttr)
	//replaceBuilder.MatchingExternal(&proto.ServerEntityPropDef{
	//	Entity:     &entityType,
	//	Attribute:  &ipAttr,
	//	UseTopoExt: &useTopoExt,
	//})

	commodities := []*proto.CommodityDTO{}
	commTypes :=  []proto.CommodityDTO_CommodityType{}
	commMetrics := metric.Metrics
	//capacity := 100.0
	for key, value := range commMetrics {
		var commType proto.CommodityDTO_CommodityType
		commType, ok := CommodityTypeMap[key]

		if !ok {
			err := fmt.Errorf("Unsupported commodity type %s", key)
			glog.Errorf(err.Error())
			continue
		}

		capacity, ok := CommodityCapMap[commType]
		if !ok {
			err := fmt.Errorf("Missing commodity capacity for type %s", commType)
			glog.Errorf(err.Error())
			continue
		}

		// TODO: Remove this if using 'millisec' unit at exporter side
		if commType == proto.CommodityDTO_RESPONSE_TIME {
			//value *= 1000 // Convert second to millisecond
		}

		// Adjust the capacity in case utilization > 1
		if value >= capacity {
			capacity = value// + 1
		}

		//if commType == proto.CommodityDTO_RESPONSE_TIME {
		//	capacity = LatencyCap
		//	value *= 1000 // Convert second to millisecond
		//	if value >= LatencyCap {
		//		//value = LatencyCap
		//		capacity = value + 1
		//	}
		//} else if commType == proto.CommodityDTO_TRANSACTION {
		//	capacity = TPSCap
		//	value = math.Min(value, TPSCap)
		//	if value >= TPSCap {
		//		//value = TPSCap - 1
		//		capacity = value + 1
		//	}
		//}

		commodity, err := builder.NewCommodityDTOBuilder(commType).
			Used(value).Capacity(capacity).Key(ip).Create()

		if err != nil {
			glog.Errorf("Error building a commodity: %s", err)
			continue
		}

		commodities = append(commodities, commodity)
		commTypes = append(commTypes, commType)
		//replaceBuilder.PatchSellingWithProperty(commType, []string{USED, CAPACITY})
	}

	//replacementMetaData := getReplacementMetaData(entityType, commTypes)

	//if len(commodities) < 2 {
	//	c, _ := builder.NewCommodityDTOBuilder(proto.CommodityDTO_RESPONSE_TIME).
	//		Used(0.0).Capacity(LatencyCap).Key(ip).Create()
	//	commodities = append(commodities, c)
	//	// TODO
	//	replaceBuilder.PatchSellingWithProperty(proto.CommodityDTO_RESPONSE_TIME, []string{USED, CAPACITY})
	//}

	id := b.getEntityId(entityType, ip)

	dto, err := builder.NewEntityDTOBuilder(entityType, id).
				DisplayName(id).
				SellsCommodities(commodities).
				WithProperty(getEntityProperty(ip)).
				ReplacedBy(getReplacementMetaData(entityType, commTypes)).
				//WithProperty(&proto.EntityDTO_EntityProperty{
				//						Namespace: &ns,
				//						Name:      &ipAttr,
				//						Value:     &ip,
				//					}).
				//ReplacedBy(replaceBuilder.Build()).
				Create()

	if err != nil {
		glog.Errorf("Error building EntityDTO from metric %v: %s", metric, err)
		return nil, err
	}

	dtos := []*proto.EntityDTO{dto}

	return dtos, nil
}

func (b *entityBuilder) getEntityId(entityType proto.EntityDTO_EntityType, entityName string) string {
	eType := proto.EntityDTO_EntityType_name[int32(entityType)]
	return fmt.Sprintf("%s-%s/%s", eType, b.scope, entityName)

}

func getReplacementMetaData(entityType proto.EntityDTO_EntityType, commTypes []proto.CommodityDTO_CommodityType) *proto.EntityDTO_ReplacementEntityMetaData {
	attr := stitchingAttr
	useTopoExt := true

	b := builder.NewReplacementEntityMetaDataBuilder().
			Matching(attr).
			MatchingExternal(&proto.ServerEntityPropDef{
						Entity:     &entityType,
						Attribute:  &attr,
						UseTopoExt: &useTopoExt,
					})

	for _, commType := range commTypes {
		b.PatchSellingWithProperty(commType, []string{USED, CAPACITY})
	}

	return b.Build()
}

func getEntityProperty(value string) *proto.EntityDTO_EntityProperty {
	attr := stitchingAttr
	ns := DefaultPropertyNamespace

	return &proto.EntityDTO_EntityProperty{
		Namespace: &ns,
		Name:      &attr,
		Value:     &value,
	}
}
