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

package common

import (
	"github.com/m3db/m3/src/cluster/client/etcd"
	"net/http"
)

// TimezoneConfig is the static config for timezone column support
type TimezoneConfig struct {
	// table to lookup timezone columns
	TableName string `yaml:"table_name"`
}

// QueryConfig is the static configuration for query.
type QueryConfig struct {
	// how much portion of the device memory we are allowed use
	DeviceMemoryUtilization float32 `yaml:"device_memory_utilization"`
	// timeout in seconds for choosing device
	DeviceChoosingTimeout int            `yaml:"device_choosing_timeout"`
	TimezoneTable         TimezoneConfig `yaml:"timezone_table"`
	EnableHashReduction   bool           `yaml:"enable_hash_reduction"`
}

// DiskStoreConfig is the static configuration for disk store.
type DiskStoreConfig struct {
	WriteSync bool `yaml:"write_sync"`
}

// HTTPConfig is the static configuration for main http server (query and schema).
type HTTPConfig struct {
	MaxConnections        int `yaml:"max_connections"`
	ReadTimeOutInSeconds  int `yaml:"read_time_out_in_seconds"`
	WriteTimeOutInSeconds int `yaml:"write_time_out_in_seconds"`
}

// ControllerConfig is the config for ares-controller client
type ControllerConfig struct {
	Address    string      `yaml:"address"`
	Headers    http.Header `yaml:"headers"`
	TimeoutSec int         `yaml:"timeout"`
}

// HeartbeatConfig is the config for timeout and check interval with etcd
type HeartbeatConfig struct {
	// heartbeat timeout value
	Timeout int `yaml:"timeout"`
	// heartbeat interval value
	Interval int `yaml:"interval"`
}

// ClusterConfig is the config for starting current instance with cluster mode
type ClusterConfig struct {
	// Enable controls whether to start in cluster mode
	Enable bool `yaml:"enable"`

	// Enable distributed mode
	Distributed bool `yaml:"distributed"`

	// Namespace is the cluster namespace to join
	Namespace string `yaml:"namespace"`

	// InstanceID is the cluster wide unique name to identify current instance
	// it can be static configured in yaml, or dynamically set on start up
	InstanceID string `yaml:"instance_id"`

	// controller config
	Controller *ControllerConfig `yaml:"controller,omitempty"`

	// etcd client required config
	Etcd etcd.Configuration `yaml:"etcd"`

	// heartbeat config
	HeartbeatConfig HeartbeatConfig `yaml:"heartbeat"`
}

// local redolog config
type DiskRedoLogConfig struct {
	// disable local disk redolog, default will be enabled
	Disabled bool `yaml:"disabled"`
}

// Kafka source config
type KafkaRedoLogConfig struct {
	// enable redolog from kafka, default will be disabled
	Enabled bool `yaml:"enabled"`
	// kafka brokers
	Brokers []string `yaml:"brokers"`
	// topic name suffix
	TopicSuffix string `yaml:"suffix"`
}

// Configs related to data import and redolog option
type RedoLogConfig struct {
	// Disk redolog config
	DiskConfig DiskRedoLogConfig `yaml:"disk"`
	// Kafka redolog config
	KafkaConfig KafkaRedoLogConfig `yaml:"kafka"`
}

// AresServerConfig is config specific for ares server.
type AresServerConfig struct {
	// HTTP port for serving.
	Port int `yaml:"port"`

	// HTTP port for debugging.
	DebugPort int `yaml:"debug_port"`

	// Directory path that stores the data and schema on local disk.
	RootPath string `yaml:"root_path"`

	// Total memory size ares can use.
	TotalMemorySize int64 `yaml:"total_memory_size"`

	// Whether to turn off scheduler.
	SchedulerOff bool `yaml:"scheduler_off"`

	// Build version of the server currently running
	Version string `yaml:"version"`

	// environment
	Env string `yaml:"env"`

	Query     QueryConfig     `yaml:"query"`
	DiskStore DiskStoreConfig `yaml:"disk_store"`
	HTTP      HTTPConfig      `yaml:"http"`
	RedoLogConfig RedoLogConfig `yaml:"redolog"`

	// Cluster determines the cluster mode configuration of aresdb
	Cluster   ClusterConfig   `yaml:"cluster"`
}
