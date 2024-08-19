package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %q\n", size, source, receivedData)

		if size < 12 {
			fmt.Println("Unexpected size:", size)
			break
		}

		receivedHeader := decodeDNSHeader([]byte(receivedData))

		// Create an empty response
		//response := []byte{}
		responseHeader := receivedHeader
		responseHeader.QueryResponseIndicator = true

		receivedQuestion := decodeQuestion([]byte(receivedData[12:]))

		response := encodeDNSHeader(responseHeader)
		response = append(response, encodeQuestion(receivedQuestion)...)

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}

type DNSHeader struct {
	PackedIdentifier       uint16
	QuestionCount          uint16
	AnswerRecordCount      uint16
	AuthorityRecordCount   uint16
	AdditionalRecordCount  uint16
	QueryResponseIndicator bool
	OperationCode          byte
	AuthoritativeAnswer    bool
	Truncation             bool
	RecursionDesired       bool
	RecursionAvailable     bool
	CheckingDisabled       bool
	AuthedData             bool
	Z                      bool
	ResponseCode           byte
}

func decodeDNSHeader(receivedData []byte) (header DNSHeader) {
	header.PackedIdentifier = binary.BigEndian.Uint16(receivedData)
	header.QuestionCount = binary.BigEndian.Uint16(receivedData[4:])
	header.AnswerRecordCount = binary.BigEndian.Uint16(receivedData[6:])
	header.AuthorityRecordCount = binary.BigEndian.Uint16(receivedData[8:])
	header.AdditionalRecordCount = binary.BigEndian.Uint16(receivedData[10:])

	flags := binary.BigEndian.Uint16(receivedData[2:])
	a := byte(flags >> 8)
	b := byte(flags & 0xFF)

	header.RecursionDesired = (a & 0x01) != 0
	header.Truncation = (a & 0x02) != 0
	header.AuthoritativeAnswer = (a & 0x04) != 0
	header.OperationCode = (a >> 3) & 0x0F
	header.QueryResponseIndicator = (a & 0x80) != 0

	header.ResponseCode = b & 0x0F
	header.CheckingDisabled = (b & 0x10) != 0
	header.AuthedData = (b & 0x20) != 0
	header.Z = (b & 0x40) != 0
	header.RecursionAvailable = (b & 0x80) != 0

	return
}

func encodeDNSHeader(header DNSHeader) (response []byte) {
	flags := uint16(0)
	if header.RecursionDesired {
		flags |= 0x0100
	}
	if header.Truncation {
		flags |= 0x0200
	}
	if header.AuthoritativeAnswer {
		flags |= 0x0400
	}
	if header.QueryResponseIndicator {
		flags |= 0x8000
	}
	flags |= uint16(header.OperationCode) << 11
	flags |= uint16(header.ResponseCode)

	if header.CheckingDisabled {
		flags |= 0x10
	}
	if header.AuthedData {
		flags |= 0x20
	}
	if header.Z {
		flags |= 0x40
	}
	if header.RecursionAvailable {
		flags |= 0x80
	}

	response = binary.BigEndian.AppendUint16(response, header.PackedIdentifier)
	response = binary.BigEndian.AppendUint16(response, flags)
	response = binary.BigEndian.AppendUint16(response, header.QuestionCount)
	response = binary.BigEndian.AppendUint16(response, header.AnswerRecordCount)
	response = binary.BigEndian.AppendUint16(response, header.AuthorityRecordCount)
	response = binary.BigEndian.AppendUint16(response, header.AdditionalRecordCount)

	return
}

type DNSQuestion struct {
	Name  []string
	Type  uint16
	Class uint16
}

func decodeQuestion(data []byte) (question DNSQuestion) {
	i := 0
	for i < len(data) {
		length := int(data[i])
		i++
		if length == 0 {
			break
		}
		question.Name = append(question.Name, string(data[i:i+length]))
		i += length
	}
	question.Type = binary.BigEndian.Uint16(data[i:])
	question.Class = binary.BigEndian.Uint16(data[i+2:])
	return
}

func encodeQuestion(question DNSQuestion) (response []byte) {
	for _, name := range question.Name {
		length := len(name)
		response = append(response, byte(length))
		response = append(response, []byte(name)...)
	}
	response = append(response, byte(0))
	response = binary.BigEndian.AppendUint16(response, question.Type)
	response = binary.BigEndian.AppendUint16(response, question.Class)
	return
}
