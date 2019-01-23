package controller

import (
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sibra"
	"github.com/scionproto/scion/go/lib/sibra/sbresv"
	"github.com/scionproto/scion/go/sibra_srv/conf"
	"github.com/scionproto/scion/go/sibra_srv/metrics"
	"github.com/scionproto/scion/go/sibra_srv/sbalgo/state"
)

/*
	This controller uses historuical data from Prometheus
	time series database to make predictions on what the
	bandwidth usage will be. Based on that information it
	creates reservations
 */
type PredictionController struct{
	// Reservation confirmation loaded from the file
	reservation *conf.Resv
	// Used for updating current steady reservation
	usage	   prometheus.Gauge
	lastBwClass sibra.BwCls
	promClient *metrics.PrometheusClient
}

func NewPredictionController(res *conf.Resv, client api.Client)(*PredictionController){
	promClient, err := metrics.NewPrometheusClient(client, conf.Get().ID)
	if err!=nil{
		log.Warn("Unable to create prometheus client")
	}
	return &PredictionController{
		reservation:res,
		usage:   metrics.SteadyPathsBandwidth.With(
			prometheus.Labels{"dstAs": res.IA.String(),
				"type":  res.PathType.String()}),
		promClient:promClient,
	}
}

func (c *PredictionController)ReservationConfirmed(resBlock *sbresv.Block){
	log.Debug("Reservation confirmed")

	prevBw := c.lastBwClass.Bps()
	currBw := resBlock.Info.BwCls.Bps()
	c.usage.Add(float64(currBw-prevBw))
	c.lastBwClass=resBlock.Info.BwCls
}

func (c *PredictionController)SetupReservation(config *conf.Conf) ReservationDetails{
	//TODO: Look at historical data to make a decision
	return ReservationDetails{}
}

func (c *PredictionController)RenewReservation(config *conf.Conf) ReservationDetails{
	//TODO: Look at historical data to make a decision
	return ReservationDetails{}
}

func (c *PredictionController)ChooseIndex([]*state.SteadyResvIdx) *state.SteadyResvIdx{
	//TODO: Look at historical data to make a decision
	return &state.SteadyResvIdx{}
}