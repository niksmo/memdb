package config

import "time"

type wal struct {
	Enabled              bool          `yaml:"enabled"`
	FlushingBatchSize    int           `yaml:"flushing_batch_size"`
	FlushingBatchTimeout time.Duration `yaml:"flushing_batch_timeout"`
	MaxSegmentSize       int           `yaml:"max_segment_size"`
	DataDirectory        string        `yaml:"data_directory"`
}

func defaultWAL() wal {
	return wal{
		Enabled:              false,
		FlushingBatchSize:    100,
		FlushingBatchTimeout: 10 * time.Millisecond,
		MaxSegmentSize:       10485760,
		DataDirectory:        "data/wal",
	}
}

type rawWAL struct {
	Enabled              *bool          `mapstructure:"enabled"`
	FlushingBatchSize    *int           `mapstructure:"flushing_batch_size"`
	FlushingBatchTimeout *time.Duration `mapstructure:"flushing_batch_timeout"`
	MaxSegmentSize       *int           `mapstructure:"max_segment_size"`
	DataDirectory        *string        `mapstructure:"data_directory"`
}

func toWAL(raw *rawWAL) wal {
	walCfg := defaultWAL()
	if raw == nil {
		return walCfg
	}

	if raw.Enabled != nil {
		walCfg.Enabled = *raw.Enabled
	}

	if raw.FlushingBatchSize != nil {
		walCfg.FlushingBatchSize = *raw.FlushingBatchSize
	}

	if raw.FlushingBatchTimeout != nil {
		walCfg.FlushingBatchTimeout = *raw.FlushingBatchTimeout
	}

	if raw.MaxSegmentSize != nil {
		walCfg.MaxSegmentSize = *raw.MaxSegmentSize
	}

	if raw.DataDirectory != nil {
		walCfg.DataDirectory = *raw.DataDirectory
	}

	return walCfg
}
