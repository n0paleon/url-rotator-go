package workerpool

import (
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Pool *ants.Pool

func IntializePool(cfg *viper.Viper, logger *zap.SugaredLogger) {
	poolSize := cfg.GetInt("task_pool.size")
	if poolSize <= 0 {
		poolSize = 10
		logger.Warnf("Invalid pool size in configuration, defaulting to %d", poolSize)
	}

	var pool, err = ants.NewPool(poolSize)
	if err != nil {
		logger.Errorf("Failed to create ants pool: %v", err)
	}

	Pool = pool

	logger.Infof("Task pool created with size %d", poolSize)
}

func ClosePool() {
	if Pool != nil {
		Pool.Release()
	}
}
