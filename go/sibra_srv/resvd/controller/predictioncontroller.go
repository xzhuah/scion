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

// This is an example function that decides how much bandwidth should be allocated to a steady reservation
func (c *PredictionController)calculateNextReservationSize(config *conf.Conf) ReservationDetails {
	reservation := ReservationDetails{
		Min:c.reservation.MinSize,
		Max:c.reservation.DesiredSize,
		Split:c.reservation.SplitCls,
		Props:c.reservation.EndProps,
		PathType:c.reservation.PathType,
	}

	// Step 1: get average usage
	avgUsage, err := c.promClient.GetAggregateForInterval(metrics.AVG, metrics.EPHEMERAL_RES_USAGE,
		c.reservation.IA.String(), c.reservation.PathType.String(),
		time.Now().Add(time.Duration(-1)*time.Minute), time.Second*10)
	if err!=nil{
		log.Warn("Unable to query Prometheus database, using default values!", "error", err)
		return reservation
	}

	// Step 2: Get information on how much bandwidth was missing
	missBw, err := c.promClient.GetChangeFrom(metrics.MISSING_BW, c.reservation.IA.String(),
		c.reservation.PathType.String(), time.Now().Sub(c.lastConfirmationTime))
	if err!=nil{
		log.Warn("Unable to get missing bandwidth for last reservation cycle", "err", err)
		return reservation
	}

	// Step 3: Get information on how much bandwidth was reserved
	if c.lastBwClass==0 {
		//TODO: Read usage from database
		log.Warn("Don't have information on last BW class")
		return reservation
	}

	utilization := avgUsage.Value/float64(c.lastBwClass.Bps())
	log.Debug("Calculating new reservation size", "averageUsage", avgUsage.Value, "missingBw", missBw.Value,
		"currRes", c.lastBwClass.Bps(), "utilization", utilization)

	if utilization<0.4 && missBw.Value<=0 {
		log.Debug("Utilization of the reservation is low, lowering it")
		reservation.Max=sibra.Bps(2*utilization*float64(c.lastBwClass.Bps())).ToBwCls(false)
	}else if missBw.Value<float64(c.lastBwClass.Bps()){
		log.Debug("Utilization we are missing some bandwidth, increasing")
		reservation.Max=sibra.Bps(float64(c.lastBwClass.Bps())+missBw.Value).ToBwCls(false)
	}else{
		log.Debug("We are missing lot of bandwidth, doubling current reservation")
		reservation.Max=(2*c.lastBwClass.Bps()).ToBwCls(false)
	}

	reservation.Max=limit(c.reservation.MinSize, c.reservation.MaxSize, sibra.Bps(avgUsage.Value*2).ToBwCls(false))
	return reservation

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
	// Don't renew too often
	if time.Now().Sub(c.lastConfirmationTime)<time.Second*30{
		return false
	}

	missBw, err := c.promClient.GetChangeFrom(metrics.MISSING_BW, c.reservation.IA.String(),
		c.reservation.PathType.String(), time.Now().Sub(c.lastConfirmationTime))
	if err!=nil{
		log.Warn("Unable to get missing bandwidth for last reservation cycle", "err", err)
		return false
	}

	// If we are missing less than 50% of bandwidht, not so important
	if missBw.Value>0.5*float64(c.lastBwClass.Bps()){
		return true
	}

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