# Meta-Proxy: Pegasus表配置中心
[![codecov](https://codecov.io/gh/pegasus-kv/meta-proxy/branch/main/graph/badge.svg?token=AGKH2FIJHR)](https://codecov.io/gh/pegasus-kv/meta-proxy)

# 概述
Meta-Proxy是Pegasus集群的表配置管理中心，客户端（兼容1.X和2.X的版本）在访问不同集群的表时不再配置不同Meta-Server地址，而是统一配置成Meta-Proxy的地址，这使得Pegasus的MetaServer变更对客户端侧完全透明。典型的应用场景包括：  
* 下线/替换Meta Server机器
* 主备集群容灾切换

# 设计
Meta-Proxy本质是Pegasus MetaServer的无状态RPC代理服务，依靠ZK存储和监听表配置的变更，并实时的反应到客户端侧。在ZK节点上存储的典型信息如下：
/ZKPathRoot/table/=>
```json
 {
  "cluster_name" : "clusterName",
  "meta_addrs" : "metaAddr1,metaAddr2,metaAddr3"
}
```
当客户端向Meta-Proxy请求某个表的信息：
* 如果本地已经缓存的有表信息或者与Meta-Server的链接，将会优先使用缓存信息；
* 如果本地缓存无对应表的信息，Meta-Proxy会从ZK上获取该表所在集群的Meta-Server地址，并把表信息和链接缓存到本地缓存中；
* 如果ZK上的表信息发声变更，Meta-Proxy会通过zk watcher监听并实时变更表信息；
* 成功获取表信息和建立与Meta-Server的连接后，向Pegasus发起请求并把结果返回给客户端。

# 使用
## 编译
```shell
make
```
随后会在bin/目录下生成`meta-proxy`的二进制文件
## 运行
```shell
./meta-proxy meta-proxy.yaml meta-proxy.log
```
以上命令表明使用`meta-proxy.yaml`文件配置启动服务，并输出日志到`meta-proxy.log`，
典型的`meta-proxy.yaml`配置如下所示：
```yaml
zookeeper:
  address: [127.0.0.1:22181,127.0.0.2:22181] # zk服务器地址
  root: /pegasus-cluster # zk节点存储表配置信息的根路径
  timeout: 1000 # ms, 连接zk节点时的超时阈值
  table_watcher_cache_capacity: 1024 # zk节点监控的最多表个数，也是meta-proxy缓存的表信息个数

metric:
  type: prometheus # 监控系统类型，同时支持“falcon”
  tags: [region=local_tst,service=meta_proxy] # 监控指标的默认tag
```
启动成功将会看到如下连接ZK的输出：
```log
2021/02/07 14:26:46 connected to [127.0.0.1:22181,127.0.0.2:22181]
2021/02/07 14:26:46 authenticated: id=321476486907785415, timeout=1000
2021/02/07 14:26:46 re-submitting `0` credentials after reconnect
```
以及日志文件的输出：
```shell
time="2021-02-07T14:26:46+08:00" level=info msg="init config: {{[127.0.0.1:22181,127.0.0.2:22181] /pegasus-cluster 1000 1024} {prometheus [region=local_tst,service=meta_proxy]}}"
time="2021-02-07T14:26:46+08:00" level=info msg="start server listen: [::]:34601"
```
## 客户端配置
客户端只需把原来的meta-server地址改配置成meta-proxy的地址即可。

**注**：更换客户端配置前，请确保Meta-Proxy连接的ZK节点已经配置好对应表的信息

# 监控
Meta-Proxy默认支持prometheus和falcon监控，并添加了三个监控指标以展示当前Meta-Proxy的服务状态：  
* client_connection_count: 记录客户端的连接数
* zk_request_count: 记录客户端的请求中从ZK上请求表信息的个数/QPS，即本地表信息缓存失效的请求数/QPS
* client_query_config_count: 客户端请求数/QPS

用户也可以根据实际服务需求配置自定义的监控指标。目前的支持的指标类型包括：count/meter、gauge

