# delay-kafka

`delay-kafka` is a delay queue service implemented in Golang using Apache Kafka. It allows you to schedule messages to be delivered at a later time.

## Features

- **Delayed Message Processing**: Schedule messages to be delivered after a specified delay.
- **High Performance**: Implemented in Golang, ensuring high performance and low latency.
- **Easy Integration**: You can produce delayed messages just like producing regular Kafka messages.
- **Horizontal Scalability**: Supports horizontal scaling to handle increased load by adding more instances.

## Requirements

- Golang 1.24.10 or later
- Apache Kafka 2.6 or later
- Redis 4.0 or later

## Installation

1. **Clone the repository**:

    ```sh
    git clone https://github.com/yvvlee/delay-kafka.git
    cd delay-kafka
    ```

2. **Build the project**:

    ```sh
    go build
    ```

3. **Run the service**:

    ```sh
    ./delay-kafka
    ```

## Configuration

Configuration is done via environment variables. Below are the environment variables you can use:

### Redis Configuration

- `DELAY_KAFKA_REDIS_ADDR`: Address of the Redis server.
- `DELAY_KAFKA_REDIS_PASSWORD`: Password for the Redis server.
- `DELAY_KAFKA_REDIS_DB`: Redis database number, default is 0.
- `DELAY_KAFKA_REDIS_READ_TIMEOUT`: Read timeout for Redis connections, default is 5s
- `DELAY_KAFKA_REDIS_WRITE_TIMEOUT`: Write timeout for Redis connections, default is 5s.
- `DELAY_KAFKA_REDIS_POOL_SIZE`: Connection pool size for Redis, default is 10.

### Kafka Configuration

- `DELAY_KAFKA_BROKERS`: Comma-separated list of Kafka brokers.
- `DELAY_KAFKA_TOPIC`: Kafka topic to store delayed messages.
- `DELAY_KAFKA_CONSUMER_GROUP`: Consumer group ID for Kafka.
- `DELAY_KAFKA_MAX_DELAY`: Maximum allowed delay. Use Go duration format such as `8760h`. The default value is `0`, which disables the limit.

## Delivery Semantics

The Kafka reader commits a consumed message before delayed-message handling completes. The Kafka writer also uses asynchronous writes. A failure after consumption can therefore lose a message. This service does not provide at-least-once or exactly-once delivery guarantees.

## Usage

- Suppose you want to produce a message with a delay of 10 seconds, the topic is "mytopic", and the payload is "hello world". The base64 encoded payload is "aGVsbG8gd29ybGQ=".
- You need to create two topics in Kafka, one for storing delayed messages (e.g., "delay_topic") and one for your custom messages ("mytopic").
- Start `delay-kafka`:
```bash
export DELAY_KAFKA_REDIS_ADDR="localhost:6379"
export DELAY_KAFKA_REDIS_PASSWORD=""
export DELAY_KAFKA_REDIS_DB="0"
export DELAY_KAFKA_REDIS_READ_TIMEOUT="10s"
export DELAY_KAFKA_REDIS_WRITE_TIMEOUT="10s"
export DELAY_KAFKA_REDIS_POOL_SIZE="10"

export DELAY_KAFKA_BROKERS="localhost:9092"
export DELAY_KAFKA_TOPIC="delay_topic"
export DELAY_KAFKA_CONSUMER_GROUP="delay-kafka-group"
export DELAY_KAFKA_MAX_DELAY="8760h"
./delay-kafka
```
- Use any Kafka client to produce a message to the "delay_topic". The message payload should be:
```json
{
   "topic": "mytopic",
   "payload": "aGVsbG8gd29ybGQ=",
   "processIn": 10,
   "toleranceSecond": 0
}
```
The detailed message definition is:

```go
type DelayMessage struct {
	// The topic of your message
	Topic string `json:"topic"`
	// Your message payload with base64 encoding
	Payload string `json:"payload"`
	// specify when to send your message relative to the current time
	ProcessIn int32 `json:"processIn"`
	// Unix timestamp in seconds for when to send your message, if ProcessIn is not set
	ProcessAt int64 `json:"processAt"`
	// ToleranceSecond, default is 0
	//	- If messageReceivedTime.Before(messageProcessTime)
	//		then enqueue the message.
	//	- If messageReceivedTime.After(messageProcessTime) and messageReceivedTime.Sub(messageProcessTime) > ToleranceSecond
	//		then ignore the message.
	//	- If messageReceivedTime.After(messageProcessTime) and messageReceivedTime.Sub(messageProcessTime) <= ToleranceSecond
	//		then process the message immediately.
	ToleranceSecond int32 `json:"toleranceSecond"`
}
```
