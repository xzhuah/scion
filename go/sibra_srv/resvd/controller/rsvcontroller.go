package controller

import (
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/sibra_srv/conf"
	"github.com/scionproto/scion/go/sibra_srv/metrics"
	"time"
)

const (
	TICK_TIME = time.Second*5
)

type ReservationManager interface {
	AddReservation(key string, reservation *conf.Resv)
	RemoveReservation(key string)
	UpdateReservation(key string, newReservation *conf.Resv)
	GetResDetails(key string)
}

type ReservationController interface {
	Start()
	Stop()
}

type BasicResController struct {
	stop     chan bool
	client 	 *metrics.PrometheusClient
}

func (c *BasicResController) Start(){
	go c.run()
}

func (c *BasicResController) Stop(){
	c.stop <- true
}

func (c *BasicResController) run() {
	client, err := metrics.NewPrometheusCli("http://localhost:9090", conf.Get().ID)
	if err!=nil{
		log.Error("Unable to create Prometheus client", "err", err)
		return
	}
	c.client=client
	c.stop = make(chan bool)
	ticker := time.NewTicker(TICK_TIME)

	//TODO: Initialize controller, load reservations

	for {
		select {
		case <-ticker.C:
			val, err := c.client.GetAggregateForInterval(metrics.AVG, time.Now().Add(time.Duration(-30)*time.Second), time.Second*25, "ff00:0:110-0")
			if err!=nil{
				log.Warn("We have problem querying prometheus", "err", err)
			}
			if val!=nil{
				log.Debug("This is the average", "avg", val.Value)
			}
		case <-c.stop:
			ticker.Stop()
			return
		}
	}
}

func (c *BasicResController) updateReservations(){
	log.Debug("Updating reservations")
}