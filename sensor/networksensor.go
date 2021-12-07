package sensor

import (
	"encoding/json"
	"fmt"

	"github.com/weaveworks/procspy"
)

type OpenPortsStatus struct {
	TcpPorts []procspy.Connection `json:"tcpPorts"`
	UdpPorts []int32              `json:"udpPorts"`
}

func SenseOpenPorts() ([]byte, error) {
	res := OpenPortsStatus{TcpPorts: make([]procspy.Connection, 0)}

	ports, err := procspy.Connections(true)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get connections :%v", err)
	}

	for c := ports.Next(); c != nil; c = ports.Next() {
		res.TcpPorts = append(res.TcpPorts, *c)
	}

	return json.Marshal(res)
}
