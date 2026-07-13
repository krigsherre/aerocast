package config

import (
	"fmt"
	"os"
	"time"

	udpTransport "github.com/krigsherre/aerocast/pkg/transport/udp"
	wsTransport "github.com/krigsherre/aerocast/pkg/transport/ws"
	"gopkg.in/yaml.v3"
)

type FileConfig struct {
	Server    ServerConfig    `yaml:"server"`
	Spatial   SpatialConfig   `yaml:"spatial"`
	Engine    EngineConfig    `yaml:"engine"`
	WebSocket WebSocketConfig `yaml:"websocket"`
	UDP       UDPConfig       `yaml:"udp"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type ServerConfig struct {
	WSListen      string `yaml:"ws_listen"`
	UDPListen     string `yaml:"udp_listen"`
	MetricsListen string `yaml:"metrics_listen"`
	MaxConns      int    `yaml:"max_connections"`
	ReadTimeout   string `yaml:"read_timeout"`
	WriteTimeout  string `yaml:"write_timeout"`
}

type SpatialConfig struct {
	GridShards int     `yaml:"grid_shards"`
	CellSizeM  float64 `yaml:"cell_size_m"`
}

type EngineConfig struct {
	TickRate        time.Duration `yaml:"-"`
	PipelineWorkers int           `yaml:"pipeline_workers"`
	MaxEntities     int           `yaml:"max_entities"`
	TickRateStr     string        `yaml:"tick_rate"`
}

type WebSocketConfig struct {
	PingInterval       string `yaml:"ping_interval"`
	PongTimeout        string `yaml:"pong_timeout"`
	MaxFrameSize       int    `yaml:"max_frame_size"`
	WriteBufferPerConn int    `yaml:"write_buffer_per_conn"`
	WriterCount        int    `yaml:"writer_count"`
}

type UDPConfig struct {
	BatchSize   int `yaml:"batch_size"`
	ReadBuffer  int `yaml:"read_buffer"`
	ChannelSize int `yaml:"channel_size"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Load(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if cfg.Engine.TickRateStr != "" {
		dur, err := time.ParseDuration(cfg.Engine.TickRateStr)
		if err != nil {
			return nil, fmt.Errorf("config: parse tick_rate: %w", err)
		}
		cfg.Engine.TickRate = dur
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *FileConfig) applyDefaults() {
	if c.Server.WSListen == "" {
		c.Server.WSListen = ":9100"
	}
	if c.Server.UDPListen == "" {
		c.Server.UDPListen = ":9101"
	}
	if c.Server.MetricsListen == "" {
		c.Server.MetricsListen = ":9102"
	}
	if c.Server.MaxConns == 0 {
		c.Server.MaxConns = 500_000
	}
	if c.Spatial.GridShards == 0 {
		c.Spatial.GridShards = 256
	}
	if c.Spatial.CellSizeM == 0 {
		c.Spatial.CellSizeM = 500
	}
	if c.Engine.TickRate == 0 {
		c.Engine.TickRate = 50 * time.Millisecond
	}
	if c.Engine.PipelineWorkers == 0 {
		c.Engine.PipelineWorkers = 4
	}
	if c.Engine.MaxEntities == 0 {
		c.Engine.MaxEntities = 2_000_000
	}
	if c.WebSocket.PingInterval == "" {
		c.WebSocket.PingInterval = "30s"
	}
	if c.WebSocket.PongTimeout == "" {
		c.WebSocket.PongTimeout = "10s"
	}
	if c.WebSocket.MaxFrameSize == 0 {
		c.WebSocket.MaxFrameSize = 1024
	}
	if c.WebSocket.WriteBufferPerConn == 0 {
		c.WebSocket.WriteBufferPerConn = 64
	}
	if c.WebSocket.WriterCount == 0 {
		c.WebSocket.WriterCount = 4
	}
	if c.UDP.BatchSize == 0 {
		c.UDP.BatchSize = 32
	}
	if c.UDP.ReadBuffer == 0 {
		c.UDP.ReadBuffer = 4 * 1024 * 1024
	}
	if c.UDP.ChannelSize == 0 {
		c.UDP.ChannelSize = 65536
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
}

func (c *FileConfig) validate() error {
	if c.Spatial.GridShards&(c.Spatial.GridShards-1) != 0 {
		return fmt.Errorf("config: grid_shards must be a power of 2, got %d", c.Spatial.GridShards)
	}
	if c.Spatial.GridShards > 65536 {
		return fmt.Errorf("config: grid_shards max is 65536, got %d", c.Spatial.GridShards)
	}
	if c.Server.MaxConns <= 0 {
		return fmt.Errorf("config: max_connections must be > 0")
	}
	if c.Engine.PipelineWorkers <= 0 {
		return fmt.Errorf("config: pipeline_workers must be > 0")
	}
	return nil
}

func (c *WebSocketConfig) ToTransportConfig() wsTransport.Config {
	return wsTransport.DefaultConfig()
}

func (c *UDPConfig) ToTransportConfig() udpTransport.Config {
	cfg := udpTransport.DefaultConfig()
	cfg.BatchSize = c.BatchSize
	cfg.ReadBuffer = c.ReadBuffer
	cfg.ChannelSize = c.ChannelSize
	return cfg
}
