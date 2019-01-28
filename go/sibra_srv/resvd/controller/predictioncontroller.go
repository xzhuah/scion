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
	"time"
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
	pathID		sibra.ID
	resKey 		string
	lastConfirmationTime time.Time
}

func NewPredictionController(res *conf.Resv, resKey string, client api.Client)(*PredictionController){
	return &PredictionController{
		reservation:res,
		usage:   metrics.SteadyPathsBandwidth.With(
			prometheus.Labels{
				"dstAs": res.IA.String(),
				"type":  res.PathType.String(),
				"resKey":resKey}),
		promClient:metrics.NewPrometheusClient(client, conf.Get().ID),
		lastConfirmationTime:time.Now(),
		resKey:resKey,
	}
}

func (c *PredictionController)ReservationConfirmed(resBlock *sbresv.Block){
	log.Debug("Reservation confirmed")

	prevBw := c.lastBwClass.Bps()
	currBw := resBlock.Info.BwCls.Bps()
	c.usage.Add(float64(currBw-prevBw))
	c.lastBwClass=resBlock.Info.BwCls

	c.lastConfirmationTime=time.Now()
}

func (c *PredictionController)SetupReservation(config *conf.Conf) ReservationDetails{
	return c.calculateNextReservationSize(config)
}

func (c *PredictionController)RenewReservation(config *conf.Conf) ReservationDetails{
	return c.calculateNextReservationSize(config)
}

func (c *PredictionController)calculateNextReservationSize(config *conf.Conf) ReservationDetails {
	reservation := ReservationDetails{
		Min:c.reservation.MinSize,
		Max:c.reservation.DesiredSize,
		Split:c.reservation.SplitCls,
		Props:c.reservation.EndProps,
		PathType:c.reservation.PathType,
	}

	avgUsage, err := c.promClient.GetAggregateForInterval(metrics.AVG, metrics.EPHEMERAL_RES_USAGE,
			time.Now().Add(time.Duration(-3)*time.Minute), time.Second*10, c.pathID.String())
	if err!=nil{
		log.Warn("Unable to query Prometheus database, using default values!", "error", err)
		return reservation
	}else{
		if avgUsage.Value<float64(c.reservation.MinSize.Bps()){
			log.Debug("There was no usage in the past, using minimum bw")
			reservation.Max=c.reservation.MinSize
		}else{
			log.Debug("There was usage in the past...")
			reservation.Max=limit(c.reservation.MinSize, c.reservation.MaxSize, sibra.Bps(avgUsage.Value*2).ToBwCls(false))
		}
		return reservation
	}
}

func (c *PredictionController)SteadyPathIDCreated(id sibra.ID){
	c.pathID=id.Copy()
}

func (c *PredictionController)ChooseIndex(pendingIndicies []*state.SteadyResvIdx) *state.SteadyResvIdx{
	minDiff := abs(c.reservation.DesiredSize.Bps() - pendingIndicies[0].Info.BwCls.Bps())
	j := 0
	for i, v := range pendingIndicies {
		diff := abs(c.reservation.DesiredSize.Bps() - v.Info.BwCls.Bps())
		if diff < minDiff {
			minDiff = diff
			j = i
		}
	}

	return pendingIndicies[j]
}

func (c *PredictionController)ShouldRenew(details ReservationDetails) bool{
	//missBw, err := c.promClient.GetChangeFrom(metrics.MISSING_BW, time.Now().Sub(c.lastConfirmationTime))
	//if err!=nil{
	//	log.Warn("Unable to get missing bandwidth for last reservation", "err", err)
	//	return false
	//}
	//
	//log.Debug("Missing bandwidth", "curr", missBw.Value)

	return false
}

func limit(min, max, value sibra.BwCls)(sibra.BwCls){
	if value>max{
		return max
	}else if value<min{
		return min
	}else{
		return value
	}
}