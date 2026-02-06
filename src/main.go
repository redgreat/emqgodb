package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/redgreat/emqgodb/src/config"
	"github.com/redgreat/emqgodb/src/mqtt"
	"github.com/redgreat/emqgodb/src/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "", "配置文件路径")
	debug := flag.Bool("debug", false, "启用调试日志")
	flag.Parse()

	// 初始化日志
	logger := initLogger(*debug)
	defer logger.Sync()

	logger.Info("EMQX to PostgreSQL 服务启动中...")

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("加载配置失败", zap.Error(err))
	}
	logger.Info("配置加载成功")

	// 初始化PostgreSQL存储
	store, err := storage.NewPostgresStorage(&cfg.PostgreSQL, logger)
	if err != nil {
		logger.Fatal("初始化PostgreSQL存储失败", zap.Error(err))
	}
	defer store.Close()

	// 初始化MQTT客户端（存储实现了MessageHandler接口）
	mqttClient, err := mqtt.NewClient(&cfg.EMQX, logger, store)
	if err != nil {
		logger.Fatal("初始化MQTT客户端失败", zap.Error(err))
	}

	// 连接MQTT
	if err := mqttClient.Connect(); err != nil {
		logger.Fatal("连接MQTT失败", zap.Error(err))
	}
	defer mqttClient.Disconnect()

	logger.Info("服务启动完成，等待消息...")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到退出信号，正在关闭服务...")
}

// initLogger 初始化日志
func initLogger(debug bool) *zap.Logger {
	var level zapcore.Level
	if debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      debug,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
