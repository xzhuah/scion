package controller

import (
	"github.com/scionproto/scion/go/lib/sibra"
	"github.com/scionproto/scion/go/lib/sibra/sbresv"
	"github.com/scionproto/scion/go/sibra_srv/conf"
	"github.com/scionproto/scion/go/sibra_srv/sbalgo/state"
)

type ReservationManager interface {
	AddReservation(key string, reservation *conf.Resv)
	RemoveReservation(key string)
	UpdateReservation(key string, newReservation *conf.Resv)
	GetResDetails(key string)
}

type ReservationDetails struct {
	Min   sibra.BwCls
	Max   sibra.BwCls
	Props sibra.EndProps
	Split sibra.SplitCls
	PathType sibra.PathType
}

type ReservationController interface {
	SetupReservation(config *conf.Conf) ReservationDetails
	RenewReservation(config *conf.Conf) ReservationDetails
	ReservationConfirmed(resBlock *sbresv.Block)
	ChooseIndex([]*state.SteadyResvIdx) *state.SteadyResvIdx
	SteadyPathIDCreated(id sibra.ID)
	ShouldRenew(details ReservationDetails) bool
}
