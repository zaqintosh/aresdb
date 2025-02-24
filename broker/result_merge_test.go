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
	"encoding/json"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber/aresdb/broker/common"
	queryCom "github.com/uber/aresdb/query/common"
	"io/ioutil"
)

var _ = ginkgo.Describe("resultMerge", func() {
	ginkgo.It("sum should work same shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123,
						"bar": 2
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Sum,
				expected: []byte(`{
					"1234": {
						"foo": 124,
						"bar": 3
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Sum,
				expected: []byte(`{}`),
			},
		})
	})

	ginkgo.It("sum should work different shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Sum,
				expected: []byte(`{
					"1234": {
						"foo": 124,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Sum,
				expected: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Sum,
				expected: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
			},
		})
	})

	ginkgo.It("count should work same shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123,
						"bar": 2
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Count,
				expected: []byte(`{
					"1234": {
						"foo": 124,
						"bar": 3
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Count,
				expected: []byte(`{}`),
			},
		})
	})

	ginkgo.It("count should work different shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Count,
				expected: []byte(`{
					"1234": {
						"foo": 124,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Count,
				expected: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Count,
				expected: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
			},
		})
	})

	ginkgo.It("max should work same shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 2,
						"bar": 1
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 2
					}
				}`),
				agg: common.Max,
				expected: []byte(`{
					"1234": {
						"foo": 2,
						"bar": 2
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Max,
				expected: []byte(`{}`),
			},
		})
	})

	ginkgo.It("max should work different shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 2
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Max,
				expected: []byte(`{
					"1234": {
						"foo": 2,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Max,
				expected: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Max,
				expected: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
			},
		})
	})

	ginkgo.It("min should work same shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 2,
						"bar": 1
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 2
					}
				}`),
				agg: common.Min,
				expected: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Min,
				expected: []byte(`{}`),
			},
		})
	})

	ginkgo.It("min should work different shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 2
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Min,
				expected: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg: common.Min,
				expected: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
			},
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Min,
				expected: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
			},
		})
	})

	ginkgo.It("avg should work same shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 2,
						"bar": 1
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 2
					}
				}`),
				agg: common.Avg,
				expected: []byte(`{
					"1234": {
						"foo": 2,
						"bar": 0.5
					}
				}`),
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{}`),
				agg:      common.Avg,
				expected: []byte(`{}`),
			},
		})
	})

	ginkgo.It("avg should error different shape", func() {
		runTests([]resultMergeTestCase{
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 2
					}
				}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg:        common.Avg,
				errPattern: "error calculating avg",
			},
			{
				lhsBytes: []byte(`{}`),
				rhsBytes: []byte(`{
					"1234": {
						"foo": 1,
						"bar": 1
					}
				}`),
				agg:        common.Avg,
				errPattern: "error calculating avg",
			},
			{
				lhsBytes: []byte(`{
					"1234": {
						"foo": 123
					}
				}`),
				rhsBytes:   []byte(`{}`),
				agg:        common.Avg,
				errPattern: "error calculating avg",
			},
		})
	})

	ginkgo.It("hll should work same shape", func() {
		data, err := ioutil.ReadFile("../testing/data/query/hll_query_results")
		Ω(err).Should(BeNil())
		lhs, _, _ := queryCom.ParseHLLQueryResults(data)
		rhs, _, _ := queryCom.ParseHLLQueryResults(data)
		ctx := newResultMergeContext(common.Hll)
		result := ctx.run(lhs[0], rhs[0])
		Ω(ctx.err).Should(BeNil())
		Ω(result).Should(Equal(lhs[0]))
	})
})

type resultMergeTestCase struct {
	lhsBytes   []byte
	rhsBytes   []byte
	agg        common.AggType
	expected   []byte
	errPattern string
}

func runTests(cases []resultMergeTestCase) {
	for _, tc := range cases {
		var lhs, rhs queryCom.AQLQueryResult
		json.Unmarshal(tc.lhsBytes, &lhs)
		json.Unmarshal(tc.rhsBytes, &rhs)
		ctx := newResultMergeContext(tc.agg)
		result := ctx.run(lhs, rhs)
		if "" == tc.errPattern {
			Ω(ctx.err).Should(BeNil())
			bs, err := json.Marshal(result)
			Ω(err).Should(BeNil())
			Ω(bs).Should(MatchJSON(tc.expected))
		} else {
			Ω(ctx.err.Error()).Should(ContainSubstring(tc.errPattern))
		}
	}
}
