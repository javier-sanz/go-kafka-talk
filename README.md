# Using Apache Kafka From Go

This repository contains the example code I used for the a talk I gave with [GoMAD](https://www.meetup.com/go-mad/)
on how to Use `Apache Kafka From Go Language`. The slides can be found [here](https://es.slideshare.net/FcoJavierSanzOlivera/kafka-from-go)

## Basic consumer and producer examples

- `Examples.java`: on this file there some examples on how to create producers in Java (no consumer though). There are synchronous and asynchronous examples. 
- `confluentExamples`: on that folder there are examples of both of consumers and producers and both asynchronous and synchronous. All them are coming from [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go)

## Run the example

In order to run the example several dependencies are needed. In order to ease that process there is a `docker-compose.yml` that starts everything needed to run the demo.
So basically the only dependency is to have docker installed locally. The images that compose will start up are:

- `wurstmeister/kafka:2.11-1.1.1`: kafka broker
- `zookeeper:3.4.11`: kafka dependency for several distributed services
- `javiersanz/go-kafka-talk:master`: this image contains all the necessary dependencies to run the consumer & producer example (go 1.11 with modules ready and librdkafka). The Dockerfile inside the project folder was the one used to create that image.
- `yandex/clickhouse-server`: a database where messages will be stored
- `spoonest/clickhouse-tabix-web-client`: a web client for Clickhouse to plot some charts
  
So you can start the demo with docker compose inside de project folder:

```console
➜  go-kafka-talk git:(master) ✗ docker-compose up -d
Creating network "go-kafka-talk_default" with the default driver
Creating go-kafka-talk_build_1 ... done
Creating go-kafka-talk_zoo1_1  ... done
Creating go-kafka-talk_kafka_1 ... done
Creating go-kafka-talk_tabix_1 ... done
Creating go-kafka-talk_clickhouse_1 ... done
```

There are to branches to use for the demo: 'autoCommit' and 'master'. Both demos are run with same commands. So regardless the branch create two docker shells as follows:

```console
go-kafka-talk git:(master) ✗ docker exec -it go-kafka-talk_build_1  bash
bash-4.4# cd /outOfGoPath/
bash-4.4#
```

Once inside go to `/outOfGoPath/` folder where the volume with the code is mounted. Now run the producer:

```console
bash-4.4# go run producerExample/producer.go kafka:9092 test
```

and

```console
go run consumerExample/consumer.go kafka:9092 testGroup test "tcp://clickhouse:9000?username=default"
```

From time to time kill the consumer and star over again. Open a browser and go to http://localhost:8081/#!/login. For connecting to clickhouse use this url http://127.0.0.1:8123 and connect using the default login (see below image) and no password.

![Clickhouse Connection](https://github.com/javier-sanz/go-kafka-talk/raw/master/tabixLogin.png)

Once inside type in the next code on the query window:

```SQL
select MessageTime, sum(MessageCount) totalMessages from default.Messages
group by MessageTime
DRAW_CHART
{
    markLine:false,
    autoAxis:true
}
```

