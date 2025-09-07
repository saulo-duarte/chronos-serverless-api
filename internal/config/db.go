package config

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(ctx context.Context, dsn string) error {
	log := WithContext(ctx)

	sqlDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.WithError(err).Errorf("Falha inicial ao abrir a conexão com o banco de dados")
		return err
	}

	db, err := sqlDB.DB()
	if err != nil {
		log.WithError(err).Errorf("Falha ao obter a instância sql.DB")
		return err
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	const maxRetries = 5
	const retryDelay = 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			DB = sqlDB
			log.Info("Conexão com o banco de dados estabelecida com sucesso")
			return nil
		}

		log.WithField("tentativa", i+1).WithError(err).Warn("Falha ao conectar com o banco de dados. Tentando novamente...")
		time.Sleep(retryDelay)
	}

	log.WithError(err).Error("Número máximo de tentativas de conexão excedido")
	return fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
}
