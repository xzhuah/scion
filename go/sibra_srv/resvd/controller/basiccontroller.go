package controller

import (
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sibra"
	"github.com/scionproto/scion/go/lib/sibra/sbresv"
	"github.com/scionproto/scion/go/sibra_srv/conf"
	"github.com/scionproto/scion/go/sibra_srv/sbalgo/state"
)
/*
	This controller implements decisions for establishing and maintaining
	steady reservations between two ASes.
	It doesn't impose any complex logic, but rather tries to maintain
	reservation bandwidht as close as possible to desired value from
	config.
 */
type BasicResController struct {
	reservation *conf.Resv
}

func NewBasicReservationController(res *conf.Resv)(ReservationController){
	return &BasicResController{
		reservation:res,
	}
}

func (c *BasicResController)SetupReservation(config *conf.Conf) ReservationDetails{
	return ReservationDetails{
		Min:c.reservation.MinSize,
		Max:c.reservation.DesiredSize,
		Split:c.reservation.SplitCls,
		Props:c.reservation.EndProps,
		PathType:c.reservation.PathType,
	}
}

func (c *BasicResController)RenewReservation(config *conf.Conf) ReservationDetails{
	return ReservationDetails{
		Min:c.reservation.MinSize,
		Max:c.reservation.DesiredSize,
		Split:c.reservation.SplitCls,
		Props:c.reservation.EndProps,
		PathType:c.reservation.PathType,
	}
}

func (c *BasicResController)ReservationConfirmed(resBlock *sbresv.Block){
	log.Debug("Reservation confirmed")
}

func (c *BasicResController)ChooseIndex(pendingIndicies []*state.SteadyResvIdx) *state.SteadyResvIdx{
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

func abs(a sibra.Bps) sibra.Bps {
	if a < 0 {
		return -a
	}
	return a
}