package test

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	proxytypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	mock "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/tests/queries/mock_client"
)

type ContractsAggregate struct {
	contractIDs      []uint64
	TwinIDs          []uint64
	NodeIDs          []uint64
	Types            []string
	States           []string
	Names            []string
	DeploymentsData  []string
	DeploymentHashes []string

	maxNumberOfPublicIPs uint64
}

const (
	CONTRACTS_TESTS = 2000
)

var contractFilterRandomValues = map[string]func(agg ContractsAggregate) interface{}{
	"ContractID": func(agg ContractsAggregate) interface{} {
		return &agg.contractIDs[rand.Intn(len(agg.contractIDs))]
	},
	"TwinID": func(agg ContractsAggregate) interface{} {
		return &agg.TwinIDs[rand.Intn(len(agg.TwinIDs))]
	},
	"NodeID": func(agg ContractsAggregate) interface{} {
		return &agg.NodeIDs[rand.Intn(len(agg.NodeIDs))]
	},
	"Type": func(agg ContractsAggregate) interface{} {
		return &agg.Types[rand.Intn(len(agg.Types))]
	},
	"State": func(agg ContractsAggregate) interface{} {
		return &agg.States[rand.Intn(len(agg.States))]
	},
	"Name": func(agg ContractsAggregate) interface{} {
		return &agg.Names[rand.Intn(len(agg.Names))]
	},
	"NumberOfPublicIps": func(agg ContractsAggregate) interface{} {
		return rndref(0, agg.maxNumberOfPublicIPs)
	},
	"DeploymentData": func(agg ContractsAggregate) interface{} {
		return &agg.DeploymentsData[rand.Intn(len(agg.DeploymentsData))]
	},
	"DeploymentHash": func(agg ContractsAggregate) interface{} {
		return &agg.DeploymentHashes[rand.Intn(len(agg.DeploymentHashes))]
	},
}

func TestContracts(t *testing.T) {
	t.Run("contracts pagination test", func(t *testing.T) {
		node := "node"
		f := proxytypes.ContractFilter{
			Type: &node,
		}

		l := proxytypes.Limit{
			Size:     5,
			Page:     1,
			RetCount: true,
		}

		for {
			want, wantCount, err := mockClient.Contracts(f, l)
			require.NoError(t, err)

			got, gotCount, err := gridProxyClient.Contracts(f, l)
			require.NoError(t, err)

			assert.Equal(t, wantCount, gotCount)

			require.True(t, reflect.DeepEqual(want, got), fmt.Sprintf("Used Filter:\n%s", serializeFilter(f)), fmt.Sprintf("Difference:\n%s", cmp.Diff(want, got)))

			if l.Page*l.Size >= uint64(wantCount) {
				break
			}
			l.Page++
		}
	})

	t.Run("contracts stress test", func(t *testing.T) {
		agg := calcContractsAggregates(&data)
		for i := 0; i < CONTRACTS_TESTS; i++ {
			l := proxytypes.Limit{
				Size:     9999999,
				Page:     1,
				RetCount: false,
			}

			f, err := randomContractsFilter(&agg)
			require.NoError(t, err)

			want, wantCount, err := mockClient.Contracts(f, l)
			require.NoError(t, err)

			got, gotCount, err := gridProxyClient.Contracts(f, l)
			require.NoError(t, err)

			assert.Equal(t, wantCount, gotCount)

			require.True(t, reflect.DeepEqual(want, got), fmt.Sprintf("Used Filter:\n%s", serializeFilter(f)), fmt.Sprintf("Difference:\n%s", cmp.Diff(want, got)))
		}
	})
}

// TestContractsFilter iterates over all ContractFilter fields, and for each one generates a random value, then runs a test between the mock client and the gridproxy client
func TestContractsFilter(t *testing.T) {
	f := proxytypes.ContractFilter{}
	fp := &f
	v := reflect.ValueOf(fp).Elem()
	l := proxytypes.Limit{
		Size:     9999999,
		Page:     1,
		RetCount: true,
	}

	agg := calcContractsAggregates(&data)

	for i := 0; i < v.NumField(); i++ {
		_, ok := contractFilterRandomValues[v.Type().Field(i).Name]
		require.True(t, ok, "Filter field %s has no random value generator", v.Type().Field(i).Name)

		randomFieldValue := contractFilterRandomValues[v.Type().Field(i).Name](agg)
		if v.Field(i).Type().Kind() != reflect.Slice {
			v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))
		}
		v.Field(i).Set(reflect.ValueOf(randomFieldValue))

		want, wantCount, err := mockClient.Contracts(f, l)
		require.NoError(t, err)

		got, gotCount, err := gridProxyClient.Contracts(f, l)
		require.NoError(t, err)

		assert.Equal(t, wantCount, gotCount)

		require.True(t, reflect.DeepEqual(want, got), fmt.Sprintf("Used Filter:\n%s", serializeFilter(f)), fmt.Sprintf("Difference:\n%s", cmp.Diff(want, got)))

		v.Field(i).Set(reflect.Zero(v.Field(i).Type()))
	}
}

func calcContractsAggregates(data *mock.DBData) (res ContractsAggregate) {
	types := make(map[string]struct{})
	for _, contract := range data.NodeContracts {
		res.contractIDs = append(res.contractIDs, contract.ContractID)
		res.maxNumberOfPublicIPs = max(res.maxNumberOfPublicIPs, contract.NumberOfPublicIPs)
		res.DeploymentsData = append(res.DeploymentsData, contract.DeploymentData)
		res.DeploymentHashes = append(res.DeploymentHashes, contract.DeploymentHash)
		res.NodeIDs = append(res.NodeIDs, contract.NodeID)
		res.States = append(res.States, contract.State)
		res.TwinIDs = append(res.TwinIDs, contract.TwinID)
		types["node"] = struct{}{}
	}

	for _, contract := range data.NameContracts {
		res.contractIDs = append(res.contractIDs, contract.ContractID)
		res.States = append(res.States, contract.State)
		res.TwinIDs = append(res.TwinIDs, contract.TwinID)
		res.Names = append(res.Names, contract.Name)
		types["name"] = struct{}{}
	}

	for _, contract := range data.RentContracts {
		res.contractIDs = append(res.contractIDs, contract.ContractID)
		res.NodeIDs = append(res.NodeIDs, contract.NodeID)
		res.States = append(res.States, contract.State)
		res.TwinIDs = append(res.TwinIDs, contract.TwinID)
		types["rent"] = struct{}{}
	}

	for typ := range types {
		res.Types = append(res.Types, typ)
	}
	sort.Slice(res.contractIDs, func(i, j int) bool {
		return res.contractIDs[i] < res.contractIDs[j]
	})
	sort.Slice(res.TwinIDs, func(i, j int) bool {
		return res.TwinIDs[i] < res.TwinIDs[j]
	})
	sort.Slice(res.NodeIDs, func(i, j int) bool {
		return res.NodeIDs[i] < res.NodeIDs[j]
	})
	sort.Slice(res.Types, func(i, j int) bool {
		return res.Types[i] < res.Types[j]
	})
	sort.Slice(res.States, func(i, j int) bool {
		return res.States[i] < res.States[j]
	})
	sort.Slice(res.Names, func(i, j int) bool {
		return res.Names[i] < res.Names[j]
	})
	sort.Slice(res.DeploymentsData, func(i, j int) bool {
		return res.DeploymentsData[i] < res.DeploymentsData[j]
	})
	sort.Slice(res.DeploymentHashes, func(i, j int) bool {
		return res.DeploymentHashes[i] < res.DeploymentHashes[j]
	})
	return
}

func randomContractsFilter(agg *ContractsAggregate) (proxytypes.ContractFilter, error) {
	f := proxytypes.ContractFilter{}
	fp := &f
	v := reflect.ValueOf(fp).Elem()

	for i := 0; i < v.NumField(); i++ {
		if rand.Float32() > .5 {
			_, ok := contractFilterRandomValues[v.Type().Field(i).Name]
			if !ok {
				return proxytypes.ContractFilter{}, fmt.Errorf("Filter field %s has no random value generator", v.Type().Field(i).Name)
			}

			randomFieldValue := contractFilterRandomValues[v.Type().Field(i).Name](*agg)
			if v.Field(i).Type().Kind() != reflect.Slice {
				v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))
			}
			v.Field(i).Set(reflect.ValueOf(randomFieldValue))
		}
	}

	return f, nil
}
