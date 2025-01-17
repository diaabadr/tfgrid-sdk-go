// Package workloads includes workloads types (vm, zdb, QSFS, public IP, gateway name, gateway fqdn, disk)
package workloads

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/gridtypes/zos"
)

// Zlog logger struct
type Zlog struct {
	Zmachine string `json:"zmachine"`
	Output   string `json:"output"`
}

// ZosWorkload generates a zlog workload
func (zlog *Zlog) ZosWorkload() gridtypes.Workload {
	url := []byte(zlog.Output)
	urlHash := md5.Sum([]byte(url))

	return gridtypes.Workload{
		Version: 0,
		Name:    gridtypes.Name(hex.EncodeToString(urlHash[:])),
		Type:    zos.ZLogsType,
		Data: gridtypes.MustMarshal(zos.ZLogs{
			ZMachine: gridtypes.Name(zlog.Zmachine),
			Output:   zlog.Output,
		}),
	}
}

func zlogs(dl *gridtypes.Deployment, name string) []Zlog {
	var res []Zlog
	for _, wl := range dl.ByType(zos.ZLogsType) {
		if !wl.Result.State.IsOkay() {
			continue
		}

		dataI, err := wl.WorkloadData()
		if err != nil {
			continue
		}

		data, ok := dataI.(*zos.ZLogs)
		if !ok {
			continue
		}

		if data.ZMachine.String() != name {
			continue
		}

		res = append(res, Zlog{
			Output:   data.Output,
			Zmachine: name,
		})
	}
	return res
}
