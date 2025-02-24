//  Copyright (c) 2017-2018 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broker

import (
	"context"
	"encoding/json"
	"github.com/uber/aresdb/broker/util"
	"github.com/uber/aresdb/cluster/topology"
	dataCli "github.com/uber/aresdb/datanode/client"
	queryCom "github.com/uber/aresdb/query/common"
	"github.com/uber/aresdb/utils"
	"net/http"
)

// StreamingScanNode implements StreamingPlanNode
type StreamingScanNode struct {
	query          queryCom.AQLQuery
	host           topology.Host
	dataNodeClient dataCli.DataNodeQueryClient
}

func (ssn *StreamingScanNode) Execute(ctx context.Context) (bs []byte, err error) {
	trial := 0
	for trial < rpcRetries {
		trial++

		var fetchErr error

		utils.GetLogger().With("host", ssn.host, "query", ssn.query).Debug("sending query to datanode")
		bs, fetchErr = ssn.dataNodeClient.QueryRaw(ctx, ssn.host, ssn.query)
		if fetchErr != nil {
			utils.GetRootReporter().GetCounter(utils.DataNodeQueryFailures).Inc(1)
			utils.GetLogger().With(
				"error", fetchErr,
				"host", ssn.host,
				"query", ssn.query,
				"trial", trial).Error("fetch from datanode failed")
			err = utils.StackError(fetchErr, "fetch from datanode failed")
			continue
		}
		utils.GetLogger().With(
			"trial", trial,
			"host", ssn.host).Info("fetch from datanode succeeded")
		break
	}
	if bs != nil {
		err = nil
	}
	return
}

func NewNonAggQueryPlan(qc *QueryContext, topo topology.Topology, client dataCli.DataNodeQueryClient, w http.ResponseWriter) (plan NonAggQueryPlan, err error) {
	headers := make([]string, len(qc.AQLQuery.Dimensions))
	for i, dim := range qc.AQLQuery.Dimensions {
		headers[i] = dim.Expr
	}
	plan.headers = headers
	plan.w = w
	plan.resultChan = make(chan streamingScanNoderesult)
	plan.limit = qc.AQLQuery.Limit

	var assignment map[topology.Host][]uint32
	assignment, err = util.CalculateShardAssignment(topo)
	if err != nil {
		return
	}

	plan.nodes = make([]*StreamingScanNode, len(assignment))
	i := 0
	for host, shards := range assignment {
		// make deep copy
		q := *qc.AQLQuery
		for _, shard := range shards {
			q.Shards = append(q.Shards, int(shard))
		}
		plan.nodes[i] = &StreamingScanNode{
			query:          q,
			host:           host,
			dataNodeClient: client,
		}
		i++
	}

	return
}

type streamingScanNoderesult struct {
	data []byte
	err  error
}

// NonAggQueryPlan implements QueryPlan
type NonAggQueryPlan struct {
	w          http.ResponseWriter
	resultChan chan streamingScanNoderesult
	headers    []string
	nodes      []*StreamingScanNode
	// number of rows needed
	limit int
	// number of rows flushed
	flushed int
}

func (nqp *NonAggQueryPlan) Execute(ctx context.Context) (err error) {
	var headersBytes []byte
	headersBytes, err = json.Marshal(nqp.headers)
	if err != nil {
		return
	}
	_, err = nqp.w.Write([]byte(`{"headers":`))
	if err != nil {
		return
	}
	_, err = nqp.w.Write(headersBytes)
	if err != nil {
		return
	}
	_, err = nqp.w.Write([]byte(`,"matrixData":[`))
	if err != nil {
		return
	}

	for _, node := range nqp.nodes {
		go func(n *StreamingScanNode) {
			var bs []byte
			bs, err = n.Execute(ctx)
			utils.GetLogger().With("dataSize", len(bs), "error", err).Debug("sending result to result channel")
			nqp.resultChan <- streamingScanNoderesult{
				data: bs,
				err:  err,
			}
		}(node)
	}

	dataNodeWaitStart := utils.Now()

	for i := 0; i < len(nqp.nodes); i++ {
		if nqp.getRowsWanted() == 0 {
			utils.GetLogger().Debug("got enough rows, exiting")
			break
		}
		res := <-nqp.resultChan

		if i == 0 {
			// only log time waited for the fastest datanode for now
			utils.GetRootReporter().GetTimer(utils.TimeWaitedForDataNode).Record(utils.Now().Sub(dataNodeWaitStart))
		}

		if res.err != nil {
			err = res.err
			return
		}
		// write rows
		if nqp.limit < 0 {
			// when no limit, flush data directly
			nqp.w.Write(res.data)
		} else {
			// with limit, we have to deserialize
			serDeStart := utils.Now()

			res.data = append([]byte("["), res.data[:]...)
			res.data = append(res.data, byte(']'))
			var resultData [][]interface{}
			err = json.Unmarshal(res.data, &resultData)

			if err != nil {
				return
			}

			if len(resultData) <= nqp.getRowsWanted() {
				nqp.w.Write(res.data[1 : len(res.data)-1])
				nqp.flushed += len(resultData)
				utils.GetLogger().With("nrows", len(resultData)).Debug("flushed batch")
			} else {
				rowsToFlush := nqp.getRowsWanted()
				dataToFlush := resultData[:rowsToFlush]
				var bs []byte
				bs, err = json.Marshal(dataToFlush)
				if err != nil {
					return
				}
				nqp.w.Write(bs[1 : len(bs)-1])
				nqp.flushed += rowsToFlush
				utils.GetLogger().With("nrows", rowsToFlush).Debug("flushed rows")
			}
			utils.GetRootReporter().GetTimer(utils.TimeSerDeDataNodeResponse).Record(utils.Now().Sub(serDeStart))
		}
		if !(nqp.getRowsWanted() == 0) && i != len(nqp.nodes)-1 {
			nqp.w.Write([]byte(`,`))
		}
	}

	_, err = nqp.w.Write([]byte(`]}`))
	return
}

func (nqp *NonAggQueryPlan) getRowsWanted() int {
	return nqp.limit - nqp.flushed
}
