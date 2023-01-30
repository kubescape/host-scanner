package sensor

import (
	"context"
	"testing"
)

func TestSenseProcSysKernel(t *testing.T) {
	_, err := SenseProcSysKernel(context.TODO())
	if err != nil {
		t.Errorf("%v", err)
	}
}
