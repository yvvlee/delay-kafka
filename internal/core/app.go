package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/yvvlee/delay-kafka/internal/config"
	"github.com/yvvlee/delay-kafka/pkg/message"
)

var ProviderSet = wire.NewSet(
	NewApp,
	NewLogger,
	NewTaskServer,
	NewTaskClient,
	NewConsumer,
	NewKafkaWriter,
	NewKafkaReader,
)

type App struct {
	taskServer *TaskServer
	consumer   *Consumer
}

func NewApp(
	taskServer *TaskServer,
	consumer *Consumer,
) *App {
	return &App{
		taskServer: taskServer,
		consumer:   consumer,
	}
}

func (a *App) Start() error {
	go a.consumer.start()
	return a.taskServer.start()
}

type TaskServer struct {
	logger *zap.Logger
	server *asynq.Server
	writer *kafka.Writer
}

func NewTaskServer(
	cfg *config.Config,
	logger *zap.Logger,
	writer *kafka.Writer,
) (*TaskServer, error) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:         cfg.Redis.Addr,
			DB:           cfg.Redis.DB,
			Password:     cfg.Redis.Password,
			PoolSize:     cfg.Redis.PoolSize,
			DialTimeout:  30 * time.Second,
			ReadTimeout:  cfg.Redis.ReadTimeout,
			WriteTimeout: cfg.Redis.WriteTimeout,
		},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// See the godoc for other configuration options
			DelayedTaskCheckInterval: time.Second,
			Logger:                   logger.Sugar(),
		},
	)
	return &TaskServer{
		logger: logger,
		writer: writer,
		server: srv,
	}, nil
}

func (d *TaskServer) start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc("main", func(ctx context.Context, task *asynq.Task) error {
		var m TaskMessage
		if err := json.Unmarshal(task.Payload(), &m); err != nil {
			d.logger.Error("failed to unmarshal task message", zap.Error(err))
			return err
		}
		value, err := m.ToKafkaMessageValue()
		if err != nil {
			d.logger.Error("failed to parse kafka message payload from task message", zap.Error(err))
			return err
		}
		d.logger.Info("received message from task message", zap.Any("message", value))
		err = d.writer.WriteMessages(ctx, kafka.Message{
			Topic:   m.Topic,
			Value:   value,
			Headers: m.ToKafkaHeaders(),
		})
		if err != nil {
			d.logger.Error("failed to write message to kafka", zap.Error(err))
		}
		return err
	})
	return d.server.Run(mux)
}

type TaskMessage struct {
	Topic   string
	Body    string
	Headers map[string]string
}

func (m *TaskMessage) ToKafkaHeaders() []kafka.Header {
	if len(m.Headers) == 0 {
		return nil
	}
	var headers []kafka.Header
	for key, value := range m.Headers {
		headers = append(headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}
	return headers
}

func (m *TaskMessage) ToKafkaMessageValue() ([]byte, error) {
	return base64.StdEncoding.DecodeString(m.Body)
}

type Consumer struct {
	logger *zap.Logger
	reader *kafka.Reader
	writer *kafka.Writer
	client *asynq.Client
}

func NewConsumer(
	logger *zap.Logger,
	reader *kafka.Reader,
	writer *kafka.Writer,
	client *asynq.Client,
) *Consumer {
	return &Consumer{
		logger: logger,
		reader: reader,
		writer: writer,
		client: client,
	}
}

func (c *Consumer) start() {
	for {
		ctx := context.Background()
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			c.logger.Error("kafka read error", zap.Error(err))
			continue
		}
		if err = c.handleMessage(ctx, &m); err != nil {
			c.logger.Error("kafka handle message error", zap.Error(err))
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, m *kafka.Message) error {
	var msg message.DelayMessage
	if err := json.Unmarshal(m.Value, &msg); err != nil {
		c.logger.Error("kafka message unmarshal error",
			zap.Error(err),
			zap.String("value", string(m.Value)),
		)
		return err
	}
	c.logger.Info("received message", zap.Any("msg", msg))
	value, err := base64.StdEncoding.DecodeString(msg.Payload)
	if err != nil {
		c.logger.Error("kafka message with invalid payload",
			zap.Error(err),
			zap.String("payload", msg.Payload),
		)
		return err
	}

	var processAt time.Time
	now := time.Now()
	if msg.ProcessIn > 0 {
		processAt = now.Add(time.Duration(msg.ProcessIn) * time.Second)
	} else if msg.ProcessAt > 0 {
		processAt = time.Unix(msg.ProcessAt, 0)
	} else {
		//process immediately
		processAt = now
	}
	if processAt.IsZero() {
		err = errors.New("the execution time is not set")
		c.logger.Error("the execution time is not set", zap.Any("msg", msg))
		return err
	}
	if processAt.Before(now.Add(time.Duration(-msg.ToleranceSecond) * time.Second)) {
		//超出容忍时间
		return nil
	}
	if !processAt.After(now) {
		//execute immediately
		err = c.writer.WriteMessages(ctx, kafka.Message{
			Topic:   m.Topic,
			Value:   value,
			Headers: m.Headers,
		})
		if err != nil {
			c.logger.Error("write kafka message error", zap.Error(err))
		}
		return err
	}
	taskMessage := &TaskMessage{
		Topic:   msg.Topic,
		Body:    msg.Payload,
		Headers: kafkaHeadersToMap(m.Headers),
	}
	payload, _ := json.Marshal(taskMessage)
	task := asynq.NewTask(
		"main",
		payload,
		asynq.TaskID(uuid.New().String()),
		asynq.ProcessAt(processAt),
	)
	if _, err = c.client.EnqueueContext(ctx, task); err != nil {
		c.logger.Error("enqueue asynq message error", zap.Error(err))
		return err
	}
	return nil
}

func NewKafkaWriter(cfg *config.Config, logger *zap.Logger) (*kafka.Writer, func()) {
	writer := &kafka.Writer{
		Addr:  kafka.TCP(cfg.Kafka.Brokers...),
		Async: true,
	}
	return writer, func() {
		if err := writer.Close(); err != nil {
			logger.Error("kafka writer close error", zap.Error(err))
		}
	}
}
func NewKafkaReader(cfg *config.Config, logger *zap.Logger) (*kafka.Reader, func()) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Kafka.Brokers,
		Topic:    cfg.Kafka.Topic,
		GroupID:  cfg.Kafka.ConsumerGroup,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return reader, func() {
		if err := reader.Close(); err != nil {
			logger.Error("kafka reader close error", zap.Error(err))
		}
	}
}

func kafkaHeadersToMap(headers []kafka.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	res := make(map[string]string)
	for _, header := range headers {
		res[header.Key] = string(header.Value)
	}
	return res
}

func NewTaskClient(cfg *config.Config, logger *zap.Logger) (*asynq.Client, func()) {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:         cfg.Redis.Addr,
		DB:           cfg.Redis.DB,
		Password:     cfg.Redis.Password,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  30 * time.Second,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	return client, func() {
		if err := client.Close(); err != nil {
			logger.Error("asynq client close error", zap.Error(err))
		}
	}
}
