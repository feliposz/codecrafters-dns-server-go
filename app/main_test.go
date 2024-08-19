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
	question, _ := decodeQuestion(data, 0)
	fmt.Printf("%+v\n", question)
}

func TestDecodePackedQuestion(t *testing.T) {
	data := []byte("\x87\xfc\x01\x00\x00\x02\x00\x00\x00\x00\x00\x00\x03abc\x11longassdomainname\x03com\x00\x00\x01\x00\x01\x03def\xc0\x10\x00\x01\x00\x01")
	question1, question1Length := decodeQuestion(data, 12)
	fmt.Printf("%+v\n", question1)
	question2, _ := decodeQuestion(data, 12+question1Length)
	fmt.Printf("%+v\n", question2)
}
