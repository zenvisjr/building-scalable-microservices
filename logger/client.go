package logger

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/zenvisjr/building-scalable-microservices/logger/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	LevelInfo  = "INFO"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
)

var (
	logInstance *Logs
	once        sync.Once
)

type Logs struct {
	Log          *zap.Logger
	service      pb.LoggerServiceClient
	conn         *grpc.ClientConn
	microservice string
}

type Config struct {
	LoggerURL string `envconfig:"LOGGER_SERVICE_URL"`
}

func InitLogger(microservice string) (*Logs, error) {
	var err error
	once.Do(func() {
		var z *zap.Logger
		z, err = zap.NewProduction()
		if err != nil {
			log.Fatalf("can't initialize zap logger: %v", err)
			return
		}
		var config Config
		if err := envconfig.Process("", &config); err != nil {
			log.Fatalf("can't process env config: %v", err)
			return
		}
		conn, e := grpc.NewClient(config.LoggerURL, grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		))
		if e != nil {
			err = e
			return
		}

		service := pb.NewLoggerServiceClient(conn)

		logInstance = &Logs{
			Log:          z,
			service:      service,
			conn:         conn,
			microservice: microservice,
		}
		z.Info("logger started")
	})
	return logInstance, err
}

// Global helper function for backward compatibility
func GetGlobalLogger() *Logs {
	if logInstance == nil {
		// debug.PrintStack()
		log.Fatal("Logger not initialized. Call InitLogger() first.")
	}
	return logInstance
}

func (l *Logs) Close() {
	if l != nil {
		l.conn.Close()
	}
}

func (l *Logs) RemoteLogs(ctx context.Context, level, msg string) {
	if l.service == nil {
		// fallback to local zap logging if remote service not ready
		l.Log.Warn("logger service is nil; falling back to local logs", zap.String("level", level), zap.String("msg", msg))
		return
	}
	method := GetMethodFromContext(ctx)
	resp, err := l.service.Log(ctx, &pb.LogRequest{
		Level:     level,
		Service:   l.microservice,
		Msg:       msg,
		Timestamp: time.Now().Format(time.RFC3339),
		Method:    method,
	})
	if err != nil {
		log.Println("Failed to log to remote service: " + err.Error())
		return
	}
	
	if !resp.Ok {
		log.Println("Logger service responded with failure")
	}
}

func (l *Logs) LocalOnlyInfo(msg string) {
	l.Log.Info(msg)
}

func (l *Logs) Info(ctx context.Context, msg string) {
	// l.Log.Info(msg)
	l.RemoteLogs(ctx, LevelInfo, msg)
}
func (l *Logs) Error(ctx context.Context, msg string) {
	// l.Log.Error(msg)
	l.RemoteLogs(ctx, LevelError, msg)
}
func (l *Logs) Fatal(ctx context.Context, msg string) {
	// l.Log.Fatal(msg)
	l.RemoteLogs(ctx, LevelFatal, msg)
}

func (l *Logs) Warn(ctx context.Context, msg string) {
	l.Log.Warn(msg)
	l.RemoteLogs(ctx, "WARN", msg)
}

func Uint64ToStr(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func IntToStr(n int) string {
	return strconv.Itoa(n)
}
