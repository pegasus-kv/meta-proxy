/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package config

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestConfig(t *testing.T) {
	Init("yaml/meta-proxy-example.yml")
	config := Configuration{
		ZookeeperOpts: zookeeperOpts{
			Address:      []string{"127.0.0.1:22181", "127.0.0.2:22181"},
			Root:         "/pegasus-cluster",
			Timeout:      1000,
			WatcherCount: 1024,
		},
		MetricsOpts: metricsOpts{
			Type: "falcon",
			Tags: []string{"region=local_tst", "service=meta_proxy"},
		},
	}

	assert.Equal(t, config, GlobalConfig)
}
