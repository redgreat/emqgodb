package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redgreat/emqgodb/src/config"
	"go.uber.org/zap"
)

// DeviceData 设备上报数据结构
type DeviceData struct {
	IMEI   string  `json:"imei"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	GpsTs  int64   `json:"gps_ts"`
	Uptime int64   `json:"uptime"`
	Csq    int16   `json:"csq"`
	Vbat   int16   `json:"vbat"`
	UpVbat int16   `json:"up_vbat"`
	IP     string  `json:"ip"`
}

// PostgresStorage PostgreSQL存储
type PostgresStorage struct {
	pool   *pgxpool.Pool
	table  string
	logger *zap.Logger
}

// NewPostgresStorage 创建PostgreSQL存储
func NewPostgresStorage(cfg *config.PostgreSQLConfig, logger *zap.Logger) (*PostgresStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL失败: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("PostgreSQL ping失败: %w", err)
	}

	s := &PostgresStorage{
		pool:   pool,
		table:  cfg.Table,
		logger: logger,
	}

	logger.Info("PostgreSQL连接成功",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database))

	return s, nil
}

// HandleMessage 实现MessageHandler接口，处理MQTT消息
func (s *PostgresStorage) HandleMessage(topic string, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 解析JSON数据
	var data DeviceData
	if err := json.Unmarshal(payload, &data); err != nil {
		s.logger.Warn("解析消息失败，跳过",
			zap.String("topic", topic),
			zap.Error(err))
		return nil
	}

	receivedAt := time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s (imei, lat, lng, gps_ts, uptime, csq, vbat, up_vbat, ip, receivetime)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, s.table)

	_, err := s.pool.Exec(ctx, query,
		data.IMEI,
		data.Lat,
		data.Lng,
		data.GpsTs,
		data.Uptime,
		data.Csq,
		data.Vbat,
		data.UpVbat,
		data.IP,
		receivedAt,
	)
	if err != nil {
		return fmt.Errorf("保存消息失败: %w", err)
	}

	s.logger.Debug("消息已保存",
		zap.String("imei", data.IMEI),
		zap.Float64("lat", data.Lat),
		zap.Float64("lng", data.Lng))

	return nil
}

// Close 关闭数据库连接
func (s *PostgresStorage) Close() {
	s.pool.Close()
	s.logger.Info("PostgreSQL连接已关闭")
}
