package addresslist

import(
	"net"
	"testing"
)

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

