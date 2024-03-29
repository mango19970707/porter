package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	kafkaWriter        *kafka.Writer
	lastWriteTime      string
	intervalDefault    time.Duration = 1
	clickhouseQueryUrl               = ""
	sql                              = ""
)

func init() {
	//os.Setenv("KAFKA_ADDR", "172.31.39.179:19092")
	//os.Setenv("KAFKA_USER", "admin")
	//os.Setenv("KAFKA_PASSWORD", "Sw@123456")
	//os.Setenv("KAFKA_TOPIC", "test1")
	//os.Setenv("CLICKHOUSE_ADDR", "172.31.39.179:8123")
	//os.Setenv("QUERY_SQL", "select ts,url,url_action,channel,id,udp,req,res from access_raw where ts > %s order by ts desc FORMAT JSON")
	// 初始化kafka写入端
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	kafkaUser := os.Getenv("KAFKA_USER")
	kafkaPwd := os.Getenv("KAFKA_PASSWORD")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaAddr == "" {
		log.Fatal("kafka address is null")
		return
	}
	var sharedTransport *kafka.Transport
	if len(kafkaUser) > 0 { // allow kafka without authorization
		mechanism, err := scram.Mechanism(scram.SHA256, kafkaUser, kafkaPwd)
		if err != nil {
			log.Fatal("kafka address is null")
			return
		}
		sharedTransport = &kafka.Transport{
			SASL: mechanism,
		}
	}
	kafkaWriter = &kafka.Writer{
		Addr:      kafka.TCP(kafkaAddr),
		Topic:     kafkaTopic,
		Transport: sharedTransport,
	}
	log.Println("kafka  address:", kafkaAddr, ", topic:", kafkaTopic)

	// 获取时间间隔
	tc := os.Getenv("WRITE_INTERVAL")
	if tc != "" {
		if t, err := strconv.Atoi(tc); err != nil {
			log.Fatal("time interval must be number")
			return
		} else {
			intervalDefault = time.Duration(t)
		}
	}
	lastWriteTime = getTenMinutesAgoTimestamp(intervalDefault)
	log.Println("write interval:", tc)

	// 获取clickhouse查询地址
	clickhouseAddr := os.Getenv("CLICKHOUSE_QUERY_URL")
	if clickhouseAddr == "" {
		log.Fatal("clickhouseAddr must not be null")
		return
	}
	clickhouseQueryUrl = fmt.Sprintf("http://%s/?password=Sw@123456&query=", clickhouseAddr)
	log.Println("clickhouse address:", clickhouseAddr)

	// 获取SQL
	sql = fmt.Sprintf(os.Getenv("QUERY_SQL"), "'"+lastWriteTime+"'")
	if sql == "" {
		log.Fatal("sql must not be null")
		return
	}
	log.Println("query sql:", sql)
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go timedWrite(sig)
	<-sig
}

func timedWrite(sign chan os.Signal) {
	sendTicker := time.NewTicker(intervalDefault * time.Minute)
	for {
		select {
		case <-sendTicker.C:
			writeData(intervalDefault)
		case <-sign:
			return
		}
	}
}

func writeData(interval time.Duration) {
	res, err := SqldbQuery(sql)
	if err != nil {
		log.Println("fail to exec sql:", err)
		return
	}
	resMap := map[string]interface{}{}
	if err = json.Unmarshal(res, &resMap); err != nil {
		log.Println("fail to json.Unmarshal res:", err)
		return
	}
	if tmp := resMap["data"]; tmp != nil {
		if data, ok := tmp.([]interface{}); ok {
			msgs := make([]kafka.Message, len(data))
			lastWriteTime = getTenMinutesAgoTimestamp(interval)
			for idx := range data {
				msgs[idx].Value, _ = json.Marshal(data[idx])
				if idx == 0 {
					var m map[string]interface{}
					if err := json.Unmarshal(msgs[idx].Value, &m); err == nil {
						if ts := m["ts"]; ts != nil {
							if tmpTs, ok := ts.(string); ok {
								lastWriteTime = tmpTs
							}
						}
					}
				}
			}
			if err = kafkaWriter.WriteMessages(context.Background(), msgs...); err != nil {
				log.Println("fail to write messages into kafka:", err)
			} else {
				log.Println("write messages into kafka successfully, length of data:", len(msgs))
			}
		}
	}
}

func SqldbQuery(q string) ([]byte, error) {
	r, err := http.Get(clickhouseQueryUrl + url.QueryEscape(q))
	if err != nil {
		log.Println("sqldbQuery error:", err)
		return nil, err
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	return body, nil
}

func getTenMinutesAgoTimestamp(internal time.Duration) string {
	tenMinutesAgo := time.Now().Add(-internal * time.Minute)
	timestamp := tenMinutesAgo.Format("2006-01-02 15:04:05")
	return timestamp
}
