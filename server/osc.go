package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type OSCMessage struct {
	Address   string
	Arguments []interface{}
}

func (msg *OSCMessage) Pack() []byte {
	buf := new(bytes.Buffer)

	var typeTag = ","
	for _, arg := range msg.Arguments {
		switch arg := arg.(type) {
		case int32:
			typeTag += "i"
			binary.Write(buf, binary.BigEndian, arg)
		case string:
			typeTag += "s"
			buf.Write(padString(arg))
		}
	}

	data := padString(msg.Address)
	data = append(data, padString(typeTag)...)
	data = append(data, buf.Bytes()...)
	return data
}

func UnpackOSC(data []byte) (*OSCMessage, error) {
	msg := &OSCMessage{}
	address, l := unpadString(data)
	msg.Address = address
	data = data[l:]

	typeTag, l := unpadString(data)
	data = data[l:]
	if typeTag == "" || typeTag[0] != ',' {
		return nil, fmt.Errorf("Invalid type tag: %s", typeTag)
	}
	typeTag = typeTag[1:]

	for _, t := range typeTag {
		switch t {
		case 'i':
			if len(data) < 4 {
				return nil, fmt.Errorf("Invalid payload")
			}
			msg.Arguments = append(msg.Arguments, int32(binary.BigEndian.Uint32(data)))
			data = data[4:]
		case 's':
			str, l := unpadString(data)
			msg.Arguments = append(msg.Arguments, str)
			data = data[l:]
		default:
			return nil, fmt.Errorf("Unsupported type tag: %c", t)
		}
	}
	return msg, nil
}

func padString(str string) []byte {
	data := make([]byte, (len(str)+4)&^3)
	copy(data, []byte(str))
	return data
}

func unpadString(data []byte) (string, int) {
	i := bytes.IndexByte(data, 0)
	if i == -1 {
		return "", 0
	}
	n := (i + 4) &^ 3
	if n > len(data) {
		return "", 0
	}
	return string(data[:i]), n
}

func RequestOSC(target string, msg OSCMessage) (*OSCMessage, error) {
	udp, err := net.Dial("udp", target)
	if err != nil {
		return nil, err
	}
	defer udp.Close()

	data := msg.Pack()
	udp.Write(data)

	buf := make([]byte, 2048)
	udp.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, err := udp.Read(buf)
	if err != nil {
		return nil, err
	}

	return UnpackOSC(buf[:n])
}
