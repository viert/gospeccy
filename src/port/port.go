package port

import (
	"log"
)

type subscriberType struct {
	tag     string
	mask    uint16
	target  uint16
	channel chan byte
}

var (
	inDataBus        []byte = make([]byte, 65536)
	outDataBus       []byte = make([]byte, 65536)
	eventSubscribers []subscriberType
)

func init() {
	for i := range inDataBus {
		inDataBus[i] = 255
		outDataBus[i] = 255
	}
}

func SetIn(portNum uint16, data byte) {
	inDataBus[portNum] = data
}

func GetIn(portNum uint16) byte {
	return inDataBus[portNum]
}

func SetOut(portNum uint16, data byte) {
	outDataBus[portNum] = data
	for _, subscriber := range eventSubscribers {
		maskedPort := portNum & subscriber.mask
		if maskedPort == subscriber.target {
			subscriber.channel <- data
		}
	}
}

func GetOut(portNum uint16) byte {
	return outDataBus[portNum]
}

func SubscribeOut(tag string, portMask uint16, portNum uint16) chan byte {
	channel := make(chan byte, 1024)
	subscriber := subscriberType{tag, portMask, portNum, channel}
	eventSubscribers = append(eventSubscribers, subscriber)
	log.Printf("%s subscribed to port %04x/%04x\n", tag, portNum, portMask)
	return channel
}
