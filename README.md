### 搬运工
定时将clickhouse数据写入Kafka

### 环境变量配置说明
| 环境变量名称          | 含义           | 备注         |
|-----------------|--------------|------------|
| KAFKA_ADDR      | Kafka地址      |            |
| KAFKA_USER      | Kafka用户名     |            |
| KAFKA_PASSWORD  | Kafka密码      |            |
| KAFKA_TOPIC     | Kafka的topic  |            |
| CLICKHOUSE_ADDR | click house地址 |            |
| QUERY_SQL       | 查询的SQL       | 必须按照时间降序排序 |
| WRITE_INTERVAL  | 查询的时间间隔    |            |

### docker-compose
```azure
  porter:
    image: docker.servicewall.cn/servicewall/porter:latest
    container_name: sw_porter
    logging: *default-logging
    environment:
      - KAFKA_ADDR=172.31.39.179:19092
      - KAFKA_USER=admin
      - KAFKA_PASSWORD=Sw@123456
      - KAFKA_TOPIC=test1
      - WRITE_INTERVAL=1
      - CLICKHOUSE_QUERY_URL=172.31.39.179:8123
      - QUERY_SQL=select ts,url,url_action,channel,id,udp,req,res from access_raw where ts > %s order by ts desc FORMAT JSON
```