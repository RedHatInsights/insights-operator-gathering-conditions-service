package kafkazerolog

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// KafkaWriter implements zerolog.LevelWriter interface and allow to use logging through Kafka
type KafkaWriter interface {
	zerolog.LevelWriter
	io.Closer
}

// KafkaLoggerConf represents the configuration for a Kafka logger using zeroconf
type KafkaLoggerConf struct {
	Broker string
	Topic  string
	Cert   string
	Level  zerolog.Level
}

type kafkaWriter struct {
	level    zerolog.Level
	producer *kafka.Writer
	topic    string
}

// NewKafkaLogger creates a new instance of kafkaWriter
func NewKafkaLogger(conf KafkaLoggerConf) (KafkaWriter, error) {
	var dialer *kafka.Dialer = nil

	if conf.Cert != "" {
		tlsConfig, err := newTLSConfig(conf.Cert)
		if err != nil {
			return nil, fmt.Errorf("SSL certificate cannot be loaded from %s. Error: %s", conf.Cert, err)
		}
		dialer = &kafka.Dialer{
			DualStack: true,
			TLS:       tlsConfig,
		}
	}

	kafkaConf := kafka.WriterConfig{
		Brokers: []string{conf.Broker},
		Topic:   conf.Topic,
		Dialer:  dialer,
		Async:   true,
	}
	return &kafkaWriter{
		level:    conf.Level,
		topic:    conf.Topic,
		producer: kafka.NewWriter(kafkaConf),
	}, nil
}

// WriteLevel implements LevelWriter interface
func (kw *kafkaWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	if level < kw.level {
		return len(p), nil
	}

	return kw.Write(p)
}

// Write implements Writer interface
func (kw *kafkaWriter) Write(p []byte) (int, error) {
	logMessage := strings.TrimSuffix(string(p), "\n")
	sendMessage := []byte(logMessage)
	if err := kw.producer.WriteMessages(context.Background(), kafka.Message{
		Value: sendMessage,
	}); err != nil {
		fmt.Println("Unable to send the log message to Kafka")
	}

	return len(p), nil
}

// Close implements Closer interface
func (kw *kafkaWriter) Close() error {
	return kw.producer.Close()
}

// newTLSConfig create the TLS configuration
func newTLSConfig(certPath string) (*tls.Config, error) {
	if certPath == "" {
		return nil, fmt.Errorf("No cert path provided. Skip")
	}
	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{},
		MinVersion:   tls.VersionTLS12,
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}
