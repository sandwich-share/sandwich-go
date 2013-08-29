package addresslist

import (
	"fmt"
	"net"
	"testing"
)

func TestHas(t *testing.T) {
	iprange := IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.255.255")}
	if !iprange.Has(net.ParseIP("129.22.0.0")) {
		t.Errorf("129.22.0.0 should be in the range")
	}
	if !iprange.Has(net.ParseIP("129.22.0.1")) {
		t.Errorf("129.22.0.1 should be in the range")
	}
	if iprange.Has(net.ParseIP("129.21.255.255")) {
		t.Errorf("129.21.255.255 should not be in the range")
	}
}

func TestRemove(t *testing.T) {
	iprangeList := make([]*IPRange, 0, 1)
	iprangeList = append(iprangeList, &IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.0.1")})
	iprangeList = append(iprangeList, &IPRange{net.ParseIP("129.22.0.3"), net.ParseIP("129.22.0.4")})
	iprangeList = append(iprangeList, &IPRange{net.ParseIP("129.22.0.6"), net.ParseIP("129.22.0.7")})
	iprangeList = append(iprangeList, &IPRange{net.ParseIP("129.22.0.9"), net.ParseIP("129.22.0.15")})
	iprangeList = remove(iprangeList, 1)
	if !iprangeList[0].Start.Equal(net.ParseIP("129.22.0.0")) || !iprangeList[1].Start.Equal(net.ParseIP("129.22.0.6")) {
		t.Errorf("did not remove second element correctly")
	}
	iprangeList = remove(iprangeList, 0)
	if !iprangeList[0].Start.Equal(net.ParseIP("129.22.0.6")) || !iprangeList[1].Start.Equal(net.ParseIP("129.22.0.9")) {
		t.Errorf("did not remove first element correctly")
	}
	iprangeList = remove(iprangeList, len(iprangeList)-1)
	if !iprangeList[0].Start.Equal(net.ParseIP("129.22.0.6")) || len(iprangeList) != 1 {
		t.Errorf("did not remove last element correctly")
	}
}

func TestInc(t *testing.T) {
	ipA := net.ParseIP("129.22.0.0")
	ipB := net.ParseIP("129.22.0.1")
	ipA = Inc(ipA)
	if !ipB.Equal(ipA) {
		t.Error("Expected: " + ipB.String() + " Got: " + ipA.String())
	}
	ipA = net.ParseIP("129.0.255.255")
	ipB = net.ParseIP("129.1.0.0")
	ipA = Inc(ipA)
	if !ipB.Equal(ipA) {
		t.Error("Expected: " + ipB.String() + " Got: " + ipA.String())
	}
	ipA = net.ParseIP("255.255.255.255")
	ipB = net.ParseIP("::1:0:0:0")
	ipA = Inc(ipA)
	if !ipB.Equal(ipA) {
		t.Error("Expected: " + ipB.String() + " Got: " + ipA.String())
	}
}

func TestShouldCombine(t *testing.T) {
	rangeA := &IPRange{net.ParseIP("129.22.0.3"), net.ParseIP("129.22.0.3")}
	rangeB := &IPRange{net.ParseIP("129.22.0.4"), net.ParseIP("129.22.0.5")}
	if !rangeA.shouldCombine(rangeB) {
		t.Errorf("Expected: true, recieved: false")
	}
	if rangeB.shouldCombine(rangeA) {
		t.Errorf("Expected: false, recieved: true")
	}
	rangeA = &IPRange{net.ParseIP("129.22.0.1"), net.ParseIP("129.22.0.4")}
	rangeB = &IPRange{net.ParseIP("129.22.0.2"), net.ParseIP("129.22.0.5")}
	if !rangeA.shouldCombine(rangeB) {
		t.Errorf("Expected: true, recieved: false")
	}
}

func TestOK(t *testing.T) {
	bwlist := NewBWList([]*IPRange{&IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.0.10")}, &IPRange{net.ParseIP("129.22.1.0"), net.ParseIP("129.22.1.10")}})
	bwlist.BlacklistRange(&IPRange{net.ParseIP("129.22.0.4"), net.ParseIP("129.22.0.5")})
	ip := net.ParseIP("129.22.0.3")
	if !bwlist.OK(ip) {
		t.Errorf(ip.String() + " should have been included")
	}
	ip = net.ParseIP("129.22.0.4")
	if bwlist.OK(ip) {
		t.Errorf(ip.String() + " should not have been included")
	}
	ip = net.ParseIP("129.22.1.0")
	if !bwlist.OK(ip) {
		t.Errorf(ip.String() + " should have been included")
	}
}

func TestBlacklistRange(t *testing.T) {
	bwlist := new(BlackWhiteList)
	shouldBe := new(BlackWhiteList)
	bwlist.BlacklistRange(&IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.0.1")})
	bwlist.BlacklistRange(&IPRange{net.ParseIP("129.22.0.3"), net.ParseIP("129.22.0.3")})
	shouldBe.Blacklist = []*IPRange{&IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.0.1")}, &IPRange{net.ParseIP("129.22.0.3"), net.ParseIP("129.22.0.3")}}
	if !bwlist.Equal(shouldBe) {
		t.Errorf("Does not add disjoint ranges correctly.\nWanted:\n%sGot:\n%s", shouldBe.String(), bwlist.String())
	}
	bwlist.BlacklistRange(&IPRange{net.ParseIP("129.22.0.4"), net.ParseIP("129.22.0.5")})
	shouldBe.Blacklist = []*IPRange{&IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.0.1")}, &IPRange{net.ParseIP("129.22.0.3"), net.ParseIP("129.22.0.5")}}
	if !bwlist.Equal(shouldBe) {
		t.Errorf("Does not add consecutive ranges correctly.\nWanted:\n%sGot:\n%s", shouldBe.String(), bwlist.String())
	}
	fmt.Println(bwlist)
	bwlist.BlacklistRange(&IPRange{net.ParseIP("129.22.0.7"), net.ParseIP("129.22.0.7")})
	fmt.Println(bwlist)
	bwlist.BlacklistRange(&IPRange{net.ParseIP("129.22.0.9"), net.ParseIP("129.22.0.10")})
	fmt.Println(bwlist)
	bwlist.BlacklistRange(&IPRange{net.ParseIP("129.22.0.2"), net.ParseIP("129.22.0.8")})
	shouldBe.Blacklist = []*IPRange{&IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.0.10")}}
	if !bwlist.Equal(shouldBe) {
		t.Errorf("Does not add multiple ranges correctly.\nWanted:\n%sGot:\n%s", shouldBe.String(), bwlist.String())
	}
}
