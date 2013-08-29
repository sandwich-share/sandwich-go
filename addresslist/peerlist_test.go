package addresslist

import (
	"net"
	"testing"
	"time"
)

func dummyItem(a, b, c, d byte) *PeerItem {
	return &PeerItem{net.IPv4(a, b, c, d), 0, makeTime(1, 1, 1)}
}

func dummyPeerList(hashes ...uint32) PeerList {
	list := make(PeerList, len(hashes))
	for _, elem := range hashes {
		list = append(list, dummyItem(0, 0, 0, byte(elem)))
	}
	return list
}

func makeTime(year, month, day int) time.Time {
	location, _ := time.LoadLocation("UTC")
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, location)
}

func TestMarshalJSON(t *testing.T) {
	location, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatalf("Location should be UTC")
	}
	item := &PeerItem{net.ParseIP("129.22.0.0"), 5, time.Date(2013, time.July, 6, 12, 0, 0, 0, location)}
	json, err := item.MarshalJSON()
	if err != nil {
		t.Fatalf(err.Error())
	}
	shouldBe := "{\"IP\":\"129.22.0.0\",\"IndexHash\":5,\"LastSeen\":\"2013-07-06T12:00:00Z\"}"
	if string(json) != shouldBe {
		t.Errorf(string(json) + " is not equal to " + shouldBe)
	}
	item = &PeerItem{net.ParseIP("129.22.1.255"), 1000000, time.Date(2013, time.July, 4, 12, 1, 40, 0, location)}
	json, err = item.MarshalJSON()
	if err != nil {
		t.Fatalf(err.Error())
	}
	shouldBe = "{\"IP\":\"129.22.1.255\",\"IndexHash\":1000000,\"LastSeen\":\"2013-07-04T12:01:40Z\"}"
	if string(json) != shouldBe {
		t.Errorf(string(json) + " is not equal to " + shouldBe)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	testTime := time.Date(2013, time.July, 6, 12, 0, 0, 0, location)
	json := []byte("{\"IP\":\"129.22.0.0\",\"IndexHash\":5,\"LastSeen\":\"2013-07-06T12:00:00Z\"}")
	item := new(PeerItem)
	err := item.UnmarshalJSON(json)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !item.IP.Equal(net.ParseIP("129.22.0.0")) || item.IndexHash != 5 || !testTime.Equal(item.LastSeen) {
		t.Errorf("item did not contain expected fields")
	}
	testTime = time.Date(2013, time.June, 6, 11, 4, 11, 0, location)
	json = []byte("{\"IP\":\"255.255.255.255\",\"IndexHash\":67667,\"LastSeen\":\"2013-06-06T11:04:11Z\"}")
	item = new(PeerItem)
	err = item.UnmarshalJSON(json)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !item.IP.Equal(net.ParseIP("255.255.255.255")) || item.IndexHash != 67667 || !testTime.Equal(item.LastSeen) {
		t.Errorf("item did not contain expected fields")
	}
}

func TestIPLess(t *testing.T) {
	ip1 := net.ParseIP("129.22.0.0")
	ip2 := net.ParseIP("129.21.255.255")
	if IPLess(ip1, ip2) {
		t.Errorf(ip1.String() + " < " + ip2.String() + "should return false, but returned true")
	}
	if !IPLess(ip2, ip1) {
		t.Errorf(ip2.String() + " < " + ip1.String() + "should return true, but returned false")
	}
	if IPLess(ip1, ip1) {
		t.Errorf(ip1.String() + " < " + ip1.String() + "should return false, but returned true")
	}
}

func TestLess(t *testing.T) {
	list := PeerList{dummyItem(255, 0, 0, 0), dummyItem(254, 0, 0, 0), dummyItem(254, 0, 0, 1)}
	if !list.Less(0, 1) {
		t.Errorf("255 is not less than 254")
	}
	if list.Less(1, 0) {
		t.Errorf("254 is less than 255")
	}
	if list.Less(1, 2) {
		t.Errorf("0 is less than 1")
	}
}

func TestSwap(t *testing.T) {
	list := dummyPeerList(1, 2, 3, 4)
	shouldBe := dummyPeerList(1, 3, 2, 4)
	list.Swap(1, 2)
	if !list.Equal(shouldBe) {
		t.Errorf("The lists should be equal")
	}
	shouldBe = dummyPeerList(4, 3, 2, 1)
	list.Swap(0, 3)
	if !list.Equal(shouldBe) {
		t.Errorf("The lists should be equal")
	}
}
