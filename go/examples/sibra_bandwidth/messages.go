package main

import "github.com/scionproto/scion/go/lib/common"

type Message interface {
	Pack(common.RawBytes) common.RawBytes
	Unpack(common.RawBytes)
	Size() int
}

const (
	TestCreate = iota
	TestCreated
	TestData
	TestEnd
	TestResult
)

// Test create message
type MsgCreateTest struct{
	Packets uint64
	Duration uint64
	Mtu 	uint64
}

func (m *MsgCreateTest)Size() int {
	return 4*8
}

func (m *MsgCreateTest)Pack(b common.RawBytes)  common.RawBytes{
	common.Order.PutUint64(b, TestCreate)
	common.Order.PutUint64(b[8:], m.Packets)
	common.Order.PutUint64(b[16:], m.Duration)
	common.Order.PutUint64(b[24:], m.Mtu)
	return b
}

func (m *MsgCreateTest)Unpack(b common.RawBytes){
	m.Packets=common.Order.Uint64(b[8:])
	m.Duration=common.Order.Uint64(b[16:])
	m.Mtu=common.Order.Uint64(b[24:])
}

// Test created reply
type MsgTestCreated struct{
	SessionId uint64
}

func (m *MsgTestCreated)Size() int {
	return 2*8
}

func (m *MsgTestCreated)Pack(b common.RawBytes)  common.RawBytes{
	common.Order.PutUint64(b, TestCreated)
	common.Order.PutUint64(b[8:], m.SessionId)
	return b
}

func (m *MsgTestCreated)Unpack(b common.RawBytes){
	m.SessionId=common.Order.Uint64(b[8:])
}

// Test data (payload)
type MsgPayload struct{
	SessionId uint64
	PacketIndex uint64
}

func (m *MsgPayload)Size() int {
	return 3*8
}

func (m *MsgPayload)Pack(b common.RawBytes)  common.RawBytes{
	common.Order.PutUint64(b, TestData)
	common.Order.PutUint64(b[8:], m.SessionId)
	common.Order.PutUint64(b[16:], m.PacketIndex)
	return b
}

func (m *MsgPayload)Unpack(b common.RawBytes){
	m.SessionId=common.Order.Uint64(b[8:])
	m.PacketIndex=common.Order.Uint64(b[16:])
}

// Test end notification
type MsgTestEnd struct{
	SessionId uint64
}

func (m *MsgTestEnd)Size() int {
	return 2*8
}

func (m *MsgTestEnd)Pack(b common.RawBytes)  common.RawBytes{
	common.Order.PutUint64(b, TestEnd)
	common.Order.PutUint64(b[8:], m.SessionId)
	return b
}

func (m *MsgTestEnd)Unpack(b common.RawBytes){
	m.SessionId=common.Order.Uint64(b[8:])
}

// Test results
type MsgTestResult struct{
	SessionId uint64
	DroppedPackets uint64
}

func (m *MsgTestResult)Size() int {
	return 3*8
}

func (m *MsgTestResult)Pack(b common.RawBytes)  common.RawBytes{
	common.Order.PutUint64(b, TestResult)
	common.Order.PutUint64(b[8:], m.SessionId)
	common.Order.PutUint64(b[16:], m.DroppedPackets)
	return b
}

func (m *MsgTestResult)Unpack(b common.RawBytes){
	m.SessionId=common.Order.Uint64(b[8:])
	m.DroppedPackets=common.Order.Uint64(b[16:])
}
