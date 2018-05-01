package monitoring

import (
	"net/http"
	"encoding/json"
	"github.com/golang/glog"
	"io/ioutil"
)

type IMetricExporter interface {
	Query() ([]*EntityMetric, error)
}

type metricExporter struct {
	endpoint string
}

func NewMetricExporter(endpoint string) *metricExporter {
	return &metricExporter{endpoint}
}

func (m *metricExporter) Query() ([]*EntityMetric, error) {
	ebytes := sendRequest(m.endpoint)

	//2. unmarshal it
	var mr MetricResponse
	if err := json.Unmarshal(ebytes, &mr); err != nil {
		glog.Errorf("Failed to un-marshal bytes: %v", string(ebytes))
		return nil, err
	}
	if mr.Status != 0 || len(mr.Data) < 1 {
		glog.Errorf("Failed to un-marshal MetricResponse: %+v", string(ebytes))
		return nil, nil
	}

	glog.V(4).Infof("mr=%+v, len=%d\n", mr, len(mr.Data))
	for i, e := range mr.Data {
		glog.V(4).Infof("[%d] %+v\n", i, e)
	}

	return mr.Data, nil
}

func sendRequest(endpoint string) []byte {
	glog.V(2).Infof("Sending request to %s", endpoint)
	resp, err := http.Get(endpoint)
	if err != nil {
		glog.Errorf("Error: %v", err)
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Error: %v", err)
		return nil
	}
	glog.V(2).Infof("Received resposne: %s", string(body))
	return body
}
