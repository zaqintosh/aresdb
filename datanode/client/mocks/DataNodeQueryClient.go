// Code generated by mockery v1.0.0
package mocks

import common "github.com/uber/aresdb/query/common"
import context "context"
import mock "github.com/stretchr/testify/mock"
import topology "github.com/uber/aresdb/cluster/topology"

// DataNodeQueryClient is an autogenerated mock type for the DataNodeQueryClient type
type DataNodeQueryClient struct {
	mock.Mock
}

// Query provides a mock function with given fields: ctx, host, query, hll
func (_m *DataNodeQueryClient) Query(ctx context.Context, host topology.Host, query common.AQLQuery, hll bool) (common.AQLQueryResult, error) {
	ret := _m.Called(ctx, host, query, hll)

	var r0 common.AQLQueryResult
	if rf, ok := ret.Get(0).(func(context.Context, topology.Host, common.AQLQuery, bool) common.AQLQueryResult); ok {
		r0 = rf(ctx, host, query, hll)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.AQLQueryResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, topology.Host, common.AQLQuery, bool) error); ok {
		r1 = rf(ctx, host, query, hll)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryRaw provides a mock function with given fields: ctx, host, query
func (_m *DataNodeQueryClient) QueryRaw(ctx context.Context, host topology.Host, query common.AQLQuery) ([]byte, error) {
	ret := _m.Called(ctx, host, query)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, topology.Host, common.AQLQuery) []byte); ok {
		r0 = rf(ctx, host, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, topology.Host, common.AQLQuery) error); ok {
		r1 = rf(ctx, host, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
