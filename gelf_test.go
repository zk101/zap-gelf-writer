package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
)

var (
	testUDPlistener *net.UDPConn
)

// TestMain
func TestMain(m *testing.M) {
	var err error

	addr := net.UDPAddr{
		Port: 1234,
		IP:   net.ParseIP("127.0.0.1"),
	}

	testUDPlistener, err = net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("UDP Listener failed: %s\n", err.Error())
	}
	defer testUDPlistener.Close()

	os.Exit(m.Run())
}

type testMsg struct {
	Host    string `json:"host"`
	Message string `json:"short_message"`
	Time    string `json:"timestamp"`
	Version string `json:"version"`
}

// testMessage
func testMessage(msg, host string) ([]byte, error) {
	testData := testMsg{
		Host:    host,
		Message: msg,
		Version: Version,
	}
	return json.Marshal(&testData)
}

// TestGELF
func TestGELF(t *testing.T) {
	packet := make([]byte, 2048)
	host := "host.example.org"
	msg := "Test Message"
	var (
		readErr error
		readLen int
	)

	testJSON, err := testMessage(msg, host)
	if err != nil {
		t.Fatalf("Getting Test Message failed: %s", err.Error())
	}

	config := DefaultConfig("127.0.0.1")
	config.Port = 1234

	gelf := New(config)

	sendLen, sendErr := gelf.Write(testJSON)
	if sendErr != nil {
		t.Fatalf("Send GELF Message failed: %s", sendErr.Error())
	}

	readLen, _, readErr = testUDPlistener.ReadFromUDP(packet)
	if readErr != nil {
		t.Fatalf("Read GELF Message failed: %s", readErr.Error())
	}

	if readLen != sendLen {
		t.Fatalf("Read Length (%d) did not equal Send Length (%d)", readLen, sendLen)
	}

	var receivedData testMsg
	if err := json.Unmarshal(packet[0:readLen], &receivedData); err != nil {
		t.Fatalf("JSON unmarshal failed: %s", err.Error())
	}

	if receivedData.Host != host {
		t.Errorf("Sent Host (%s) does not equal Received Host (%s)", host, receivedData.Host)
	}

	if receivedData.Message != msg {
		t.Errorf("Sent Message (%s) does not equal Received Message (%s)", msg, receivedData.Message)
	}

	if receivedData.Version != Version {
		t.Errorf("Sent Version (%s) does not equal Received Version (%s)", Version, receivedData.Version)
	}
}

// TestGELFgzip
func TestGELFgzip(t *testing.T) {
	packet := make([]byte, 2048)
	host := "host.example.org"
	msg := "Test Message"
	var (
		readErr error
		readLen int
	)

	testJSON, err := testMessage(msg, host)
	if err != nil {
		t.Fatalf("Getting Test Message failed: %s", err.Error())
	}

	config := DefaultConfig("127.0.0.1")
	config.Port = 1234
	config.Compression = CompressionGZip

	gelf := New(config)

	_, sendErr := gelf.Write(testJSON)
	if sendErr != nil {
		t.Fatalf("Send GELF Message failed: %s", sendErr.Error())
	}

	readLen, _, readErr = testUDPlistener.ReadFromUDP(packet)
	if readErr != nil {
		t.Fatalf("Read GELF Message failed: %s", readErr.Error())
	}

	packetUnpack, err := gzip.NewReader(bytes.NewBuffer(packet[0:readLen]))
	if err != nil {
		t.Fatalf("GZip NewReader failed: %s", err.Error())
	}
	defer packetUnpack.Close()

	packetData, err := ioutil.ReadAll(packetUnpack)
	if err != nil {
		t.Fatalf("ReadAll failed: %s", err.Error())
	}

	var receivedData testMsg
	if err := json.Unmarshal(packetData, &receivedData); err != nil {
		t.Fatalf("JSON unmarshal failed: %s", err.Error())
	}

	if receivedData.Host != host {
		t.Errorf("Sent Host (%s) does not equal Received Host (%s)", host, receivedData.Host)
	}

	if receivedData.Message != msg {
		t.Errorf("Sent Message (%s) does not equal Received Message (%s)", msg, receivedData.Message)
	}

	if receivedData.Version != Version {
		t.Errorf("Sent Version (%s) does not equal Received Version (%s)", Version, receivedData.Version)
	}
}

// TestGELFzlib
func TestGELFzlib(t *testing.T) {
	packet := make([]byte, 2048)
	host := "host.example.org"
	msg := "Test Message"
	var (
		readErr error
		readLen int
	)

	testJSON, err := testMessage(msg, host)
	if err != nil {
		t.Fatalf("Getting Test Message failed: %s", err.Error())
	}

	config := DefaultConfig("127.0.0.1")
	config.Port = 1234
	config.Compression = CompressionZLib

	gelf := New(config)

	_, sendErr := gelf.Write(testJSON)
	if sendErr != nil {
		t.Fatalf("Send GELF Message failed: %s", sendErr.Error())
	}

	readLen, _, readErr = testUDPlistener.ReadFromUDP(packet)
	if readErr != nil {
		t.Fatalf("Read GELF Message failed: %s", readErr.Error())
	}

	packetUnpack, err := zlib.NewReader(bytes.NewBuffer(packet[0:readLen]))
	if err != nil {
		t.Fatalf("GZip NewReader failed: %s", err.Error())
	}
	defer packetUnpack.Close()

	packetData, err := ioutil.ReadAll(packetUnpack)
	if err != nil {
		t.Fatalf("ReadAll failed: %s", err.Error())
	}

	var receivedData testMsg
	if err := json.Unmarshal(packetData, &receivedData); err != nil {
		t.Fatalf("JSON unmarshal failed: %s", err.Error())
	}

	if receivedData.Host != host {
		t.Errorf("Sent Host (%s) does not equal Received Host (%s)", host, receivedData.Host)
	}

	if receivedData.Message != msg {
		t.Errorf("Sent Message (%s) does not equal Received Message (%s)", msg, receivedData.Message)
	}

	if receivedData.Version != Version {
		t.Errorf("Sent Version (%s) does not equal Received Version (%s)", Version, receivedData.Version)
	}
}

// TestGELFchunked
func TestGELFchunked(t *testing.T) {
	packetOne := make([]byte, 2048)
	packetTwo := make([]byte, 2048)
	host := "host.example.org"
	msg := "Test Message"
	var (
		readErr    error
		readLenOne int
		readLenTwo int
	)

	testJSON, err := testMessage(msg, host)
	if err != nil {
		t.Fatalf("Getting Test Message failed: %s", err.Error())
	}

	config := DefaultConfig("127.0.0.1")
	config.Port = 1234
	config.MaxChunkSize = 80

	gelf := New(config)

	_, sendErr := gelf.Write(testJSON)
	if sendErr != nil {
		t.Fatalf("Send GELF Message failed: %s", sendErr.Error())
	}

	readLenOne, _, readErr = testUDPlistener.ReadFromUDP(packetOne)
	if readErr != nil {
		t.Fatalf("Read GELF packet one failed: %s", readErr.Error())
	}

	readLenTwo, _, readErr = testUDPlistener.ReadFromUDP(packetTwo)
	if readErr != nil {
		t.Fatalf("Read GELF packet two failed: %s", readErr.Error())
	}

	packet := packetOne[12:readLenOne]
	packet = append(packet, packetTwo[12:readLenTwo]...)

	var receivedData testMsg
	if err := json.Unmarshal(packet, &receivedData); err != nil {
		t.Fatalf("JSON unmarshal failed: %s", err.Error())
	}

	if receivedData.Host != host {
		t.Errorf("Sent Host (%s) does not equal Received Host (%s)", host, receivedData.Host)
	}

	if receivedData.Message != msg {
		t.Errorf("Sent Message (%s) does not equal Received Message (%s)", msg, receivedData.Message)
	}

	if receivedData.Version != Version {
		t.Errorf("Sent Version (%s) does not equal Received Version (%s)", Version, receivedData.Version)
	}

	// Header Checks - Magic Numbers
	if int(packetOne[0]) != 30 {
		t.Errorf("Packet One first magic byte (%d) not equal 30", int(packetOne[0]))
	}

	if int(packetOne[1]) != 15 {
		t.Errorf("Packet One first magic byte (%d) not equal 15", int(packetOne[1]))
	}

	if int(packetTwo[0]) != 30 {
		t.Errorf("Packet Two first magic byte (%d) not equal 30", int(packetTwo[0]))
	}

	if int(packetTwo[1]) != 15 {
		t.Errorf("Packet Two first magic byte (%d) not equal 15", int(packetTwo[1]))
	}

	// Header Checks - Message ID
	for x := 2; x < 10; x++ {
		if packetOne[x] != packetTwo[x] {
			t.Errorf("Packet Message ID did not match (%d) != (%d)", int(packetOne[x]), int(packetTwo[x]))
		}
	}

	// Header Checks - Sequence Number (start 0)
	if int(packetOne[10]) != 0 {
		t.Errorf("Packet One sequence number (%d) not equal 0", int(packetOne[10]))
	}

	if int(packetTwo[10]) != 1 {
		t.Errorf("Packet Two sequence number (%d) not equal 1", int(packetTwo[10]))
	}

	// Header Checks - Sequence Count
	if int(packetOne[11]) != 2 {
		t.Errorf("Packet One sequence count (%d) not equal 0", int(packetOne[11]))
	}

	if int(packetTwo[11]) != 2 {
		t.Errorf("Packet Two sequence count (%d) not equal 1", int(packetTwo[11]))
	}
}

// EOF
