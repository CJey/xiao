package netkit

import (
	"testing"
)

func TestMyManIP(t *testing.T) {
	var ip = MyManIP()
	t.Log("Man ip got: " + ip.String())
}

func TestMyWanIP(t *testing.T) {
	var ip = MyWanIP()
	t.Log("Wan ip got: " + ip.String())
}
