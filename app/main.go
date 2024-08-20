package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func main() {
	var resolverAddr string
	flag.StringVar(&resolverAddr, "resolver", "", "set resolver address")
	flag.Parse()

	if !flag.Parsed() {
		flag.Usage()
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	var fwdResolver *net.Resolver
	if resolverAddr != "" {
		fmt.Println("connecting to ", resolverAddr)
		fwdResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout: time.Second * time.Duration(10),
				}
				return dialer.DialContext(ctx, network, resolverAddr)
			},
		}
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

		responseHeader := receivedHeader
		responseHeader.QueryResponseIndicator = true
		if receivedHeader.OperationCode == 0 {
			responseHeader.ResponseCode = 0
		} else {
			responseHeader.ResponseCode = 4
		}

		offset := 12 // skipping header

		encodedQuestions := []byte{}
		encodedAnswers := []byte{}

		// log.Println("receivedHeader.QuestionCount", receivedHeader.QuestionCount)
		for i := 0; i < int(receivedHeader.QuestionCount); i++ {
			// log.Println("question #", i)
			receivedQuestion, questionLength := decodeQuestion([]byte(receivedData), offset)
			// log.Printf("%+v\n", receivedQuestion)

			answer := DNSAnswer{
				Name:  receivedQuestion.Name,
				Type:  receivedQuestion.Type,
				Class: receivedQuestion.Class,
				TTL:   60,
				Data:  []byte{8, 8, 8, 8},
			}
			// log.Printf("%+v\n", answer)

			encodedQuestion := encodeQuestion(receivedQuestion)
			encodedQuestions = append(encodedQuestions, encodedQuestion...)

			if fwdResolver != nil {
				host := strings.Join(receivedQuestion.Name, ".")

				log.Println("forward request for", host)
				ips, err := fwdResolver.LookupIP(context.Background(), "ip4", host)
				if err == nil {
					log.Println("Got", ips)
					for _, ip := range ips {
						answer.Data = ip.To4()
						encodedAnswer := encodeAnswer(answer)
						encodedAnswers = append(encodedAnswers, encodedAnswer...)
						responseHeader.AnswerRecordCount++
					}
				} else {
					log.Println(err)
					answer.Data = []byte{8, 8, 8, 8}
					encodedAnswer := encodeAnswer(answer)
					encodedAnswers = append(encodedAnswers, encodedAnswer...)
					responseHeader.AnswerRecordCount++
					continue
				}
			}

			offset += questionLength
		}

		encodedHeader := encodeDNSHeader(responseHeader)

		response := []byte{}
		response = append(response, encodedHeader...)
		response = append(response, encodedQuestions...)
		response = append(response, encodedAnswers...)

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

func decodeQuestion(data []byte, offset int) (question DNSQuestion, size int) {
	i := offset
	for i < len(data) {
		length := int(data[i])
		i++
		if length == 0 {
			break
		}
		if length <= 63 {
			question.Name = append(question.Name, string(data[i:i+length]))
			i += length
		} else {
			// log.Println("Found pointer at ", i)
			compressedOffset := int(binary.BigEndian.Uint16(data[i-1:]) & 0b0011111111111111)
			j := compressedOffset
			for j < len(data) {
				length := int(data[j])
				// log.Println("j=", j)
				// log.Println("length=", length)
				j++
				if length == 0 {
					break
				}
				temp := string(data[j : j+length])
				// log.Println("temp=", temp)
				question.Name = append(question.Name, temp)
				j += length
			}
			i++
			break
		}
	}
	question.Type = binary.BigEndian.Uint16(data[i:])
	i += 2
	question.Class = binary.BigEndian.Uint16(data[i:])
	i += 2
	size = i - offset
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

type DNSAnswer struct {
	Name  []string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  []byte
}

func encodeAnswer(answer DNSAnswer) (response []byte) {
	for _, name := range answer.Name {
		length := len(name)
		response = append(response, byte(length))
		response = append(response, []byte(name)...)
	}
	response = append(response, byte(0))
	response = binary.BigEndian.AppendUint16(response, answer.Type)
	response = binary.BigEndian.AppendUint16(response, answer.Class)
	response = binary.BigEndian.AppendUint32(response, answer.TTL)
	response = binary.BigEndian.AppendUint16(response, uint16(len(answer.Data)))
	response = append(response, answer.Data...)
	return
}
