package main

import (
	"fmt"
	"testing"
)

func TestDecodeDNSHeader(t *testing.T) {
	data := []byte{0x86, 0x2a, 0x81, 0x80, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
	header := decodeDNSHeader(data)
	//fmt.Printf("%+v\n", header)
	fmt.Println("PackedIdentifier: ", header.PackedIdentifier)
	fmt.Println("QueryResponseIndicator: ", header.QueryResponseIndicator)
	fmt.Println("OperationCode: ", header.OperationCode)
	fmt.Println("AuthoritativeAnswer: ", header.AuthoritativeAnswer)
	fmt.Println("Truncation: ", header.Truncation)
	fmt.Println("RecursionDesired: ", header.RecursionDesired)
	fmt.Println("RecursionAvailable: ", header.RecursionAvailable)
	fmt.Println("CheckingDisabled: ", header.CheckingDisabled)
	fmt.Println("AuthedData: ", header.AuthedData)
	fmt.Println("Z: ", header.Z)
	fmt.Println("ResponseCode: ", header.ResponseCode)
	fmt.Println("QuestionCount: ", header.QuestionCount)
	fmt.Println("AnswerRecordCount: ", header.AnswerRecordCount)
	fmt.Println("AuthorityRecordCount: ", header.AuthorityRecordCount)
	fmt.Println("AdditionalRecordCount: ", header.AdditionalRecordCount)
}

func TestEncodeDecode(t *testing.T) {
	orig := DNSHeader{
		PackedIdentifier:       34346,
		RecursionDesired:       true,
		Truncation:             false,
		AuthoritativeAnswer:    false,
		OperationCode:          0,
		QueryResponseIndicator: true,
		ResponseCode:           0,
		CheckingDisabled:       false,
		AuthedData:             false,
		Z:                      false,
		RecursionAvailable:     true,
		QuestionCount:          1,
		AnswerRecordCount:      1,
		AuthorityRecordCount:   0,
		AdditionalRecordCount:  0,
	}
	data := encodeDNSHeader(orig)
	result := decodeDNSHeader(data)
	if orig != result {
		t.Fail()
	}
}

func TestDecodeQuestion(t *testing.T) {
	data := []byte{0x06, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01}
	question := decodeQuestion(data)
	fmt.Printf("%+v\n", question)
}
