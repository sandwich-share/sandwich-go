package addresslist

import(
	"testing"
	"net"
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

func Testremove(t *testing.T) {
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
	iprangeList = remove(iprangeList, len(iprangeList) - 1)
	if !iprangeList[0].Start.Equal(net.ParseIP("129.22.0.6")) || len(iprangeList) != 1 {
		t.Errorf("did not remove last element correctly")
	}
}

