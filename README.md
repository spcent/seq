# seq
基于mysql的全局序列号生成器，用go实现，同时支持worker和db模式。对于订单号，可以选择worker模式，对于用户id这种，可以采用db模式。

### 特性

* 分布式：可任意横向扩展
* 高性能：分配ID只能访问内存
* 易用性：对外提供HTTP服务
* 唯一性：MySQL自增ID，永不重复
* 高可靠：MySQL持久化

### 依赖项

本项目使用下列优秀的项目作为必要组件。

* gopkg.in/yaml.v2
* github.com/go-sql-driver/mysql
* github.com/satori/go.uuid

### 安装

**注意:需要在启动之前创建数据库并修改配置文件中数据库的配置。**

单独编译：

```shell
git clone https://github.com/spcent/seq.git
cd seq
go build .
./seq
```
Docker 方式：

Dockerfile 使用了 Docker 多阶段构建功能，需保证 Docker 版本在 17.05 及以上。详见：[Use multi-stage builds](https://docs.docker.com/develop/develop-images/multistage-build/)

```shell
git clone https://github.com/spcent/seq.git
cd seq
docker build seq:latest .
docker run -p 8000:8000 seq:latest
```

### 初始化数据库

数据库名称可以自定义，修改config.yml即可。
然后导入下面的SQL生成数据表。

```mysql
create database seq;

CREATE TABLE `seq_number` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `uuid` char(36) NOT NULL COMMENT '机器识别码',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_uuid` (`uuid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
```

### 使用

```shell
curl http://localhost:8000/nextId
{"code":0,"msg":"ok","data":{"id":101}}

curl http://localhost:8000/nextIdSimple
102

curl http://localhost:8000/worker/1
{"code":0,"msg":"ok","data":{"id":390637407633936384}}
```

### 原理

本项目设计原理来自 携程技术中心 的[干货 | 分布式架构系统生成全局唯一序列号的一个思路](https://mp.weixin.qq.com/s/F7WTNeC3OUr76sZARtqRjw)。

服务初始化后第一次请求会在 MySQL 数据库中插入一条数据，以生成初始 ID。

后续的请求，都会在内存中进行自增返回，并且保证返回的 ID 不会超过设置的上限，到达上限后会再次从 MySQL 中更新数据，返回新的初始 ID 。

### 参考
* seqsrv（https://gitee.com/qichengzx/seqsvr）: 全局唯一序列号生成服务
* Tinyid（https://github.com/didi/tinyid）: 是用Java开发的一款分布式id生成系统，基于数据库号段算法实现，关于这个算法可以参考美团leaf或者tinyid原理介绍。Tinyid扩展了leaf-segment算法，支持了多db(master)，同时提供了java-client(sdk)使id生成本地化，获得了更好的性能与可用性。Tinyid在滴滴客服部门使用，均通过tinyid-client方式接入，每天生成亿级别的id。
* twitter snowflake（https://github.com/twitter-archive/snowflake）
* 百度uid-generator（https://github.com/baidu/uid-generator）: 这是基于snowflake方案实现的开源组件，借用未来时间、缓存等手段，qps可达600w+
* 美团leaf（https://tech.meituan.com/MT_Leaf.html）: 该篇文章详细的介绍了db号段和snowflake方案，近期也进行了Leaf开源

##### 核心SQL

```mysql
REPLACE INTO `seq_number` (uuid) VALUES ("54f5a3e2-e04c-4664-81db-d7f6a1259d01");
```

### TODO
* 高可用（在ab测试中，发现存在请求被hang住的情况，响应时间不是太稳定）
* 批量获取
