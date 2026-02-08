package postgresql

import (
	"log"

	"github.com/herobeniyoutube/vk-forwarder/domain"
)

func (p *PostgresDb) AddIdempotencyKey(key string) error {
	err := p.db.Table("idempotency_keys").Create(&domain.IdempotencyKey{Key: key}).Error
	if err == nil {
		log.Printf("Added idempotency key %s", key)
	}
	return err
}

func (p *PostgresDb) HasIdempotencyKey(key string) (bool, error) {
	var count int64
	if err := p.db.Table("idempotency_keys").
		Where("key = ?", key).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
