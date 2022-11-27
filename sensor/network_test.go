package sensor

import "testing"

func TestSenseOpenPorts(t *testing.T) {
	_, err := SenseOpenPorts()
	if err != nil {
		t.Errorf("%v", err)
	}
}
