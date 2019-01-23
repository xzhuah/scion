package metrics

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"time"
)

const (
	REQ_TIMEOUT = time.Millisecond*300
)

type PrometheusClient struct {
	api 	 v1.API
	id 		 string
}

func NewPrometheusClient(client api.Client, serviceId string)(*PrometheusClient, error){
	cli := &PrometheusClient{
		api:v1.NewAPI(client),
		id:serviceId,
	}
	return cli, nil
}

type TimeSeries struct{
	Samples 	[]model.SamplePair
}

type AgregateScalar struct{
	Value 		float64
}

type AggregateFunction int
const(
	AVG AggregateFunction = iota
	SUM
	MIN
	MAX
)

var aggregate_functions = map[AggregateFunction] string {
	AVG:"avg",
	SUM:"sum",
	MIN:"min",
	MAX:"max",
}

func (c *PrometheusClient)GetAggregateForInterval(f AggregateFunction, when time.Time, duration time.Duration, steadyPathId string)(*AgregateScalar, error){
	ctx, cancelF := context.WithTimeout(context.Background(), REQ_TIMEOUT)
	defer cancelF()

	queryString := fmt.Sprintf("%s_over_time(%s_%s{elem=\"%s\"}[%ds])",aggregate_functions[f],
			NAMESPACE, EPHEMERAL_RES_USAGE, c.id, int(duration.Seconds()))
	val, err := c.api.Query(ctx, queryString, when)
	if err!=nil {
		log.Debug("Error processing", "query_string", queryString)
		return nil, err
	}
	log.Debug("Got results", "type", val.Type(), "val", val.String())
	results := val.(model.Vector)
	if (len(results)>0){
		return &AgregateScalar{
			Value:float64(results[0].Value),
		}, nil
	}

	return nil, nil
}

func (c *PrometheusClient)GetEphResTimestamps(from time.Time, duration time.Duration, steadyPathId string)(*TimeSeries, error){
	ctx, cancelF := context.WithTimeout(context.Background(), REQ_TIMEOUT)
	defer cancelF()

	queryString := fmt.Sprintf("%s_%s{elem=\"%s\"}",NAMESPACE, EPHEMERAL_RES_USAGE, c.id)
	val, err := c.api.QueryRange(ctx, queryString,
		v1.Range{
			Start:from,
			End:from.Add(duration),
			Step:time.Second*5,
		},
	)
	if err!=nil{
		return nil, err
	}

	if val.Type()!=model.ValMatrix{
		return nil, common.NewBasicError("Unexpected value type. Expecting val type matrix", nil)
	}
	results:=val.(model.Matrix)
	if len(results)>1{
		return nil, common.NewBasicError("Unexpected size of the vector, expected size=1", nil)
	}

	for _,v := range results{
		res:=&TimeSeries{Samples:v.Values}
		return res, nil
	}

	return nil, nil
}
