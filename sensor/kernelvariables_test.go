package sensor

import "testing"

func TestSenseProcSysKernel(t *testing.T) {
	_, err := SenseProcSysKernel()
	if err != nil {
		t.Errorf("%v", err)
	}
}
