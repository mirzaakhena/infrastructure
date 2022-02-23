package database

import (
	"fmt"
	"github.com/go-bongo/bongo"
	"infrastructure/shared/infrastructure/config"
)

func NewMongoDBConnection(cfg *config.Config) *bongo.Connection {

	cs := cfg.Database.Host

	if cfg.Database.Username != "" {
		cs = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Database)
	}

	bongoCfg := &bongo.Config{
		ConnectionString: cs,
		Database:         cfg.Database.Database,
	}
	
	connection, err := bongo.Connect(bongoCfg)
	if err != nil {
		panic(err)
	}

	return connection
}
