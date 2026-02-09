package postgresql

import (
	"log"

	"github.com/herobeniyoutube/vk-forwarder/application"
	"github.com/herobeniyoutube/vk-forwarder/domain"
	"github.com/herobeniyoutube/vk-forwarder/domain/statuses"
	"gorm.io/gorm/clause"
)

func (p *PostgresDb) AddOrUpdateIdempotencyKey(key string, status statuses.IdempotencyStatus) (*statuses.IdempotencyStatus, error) {
	entity := &domain.IdempotencyKey{
		Key:    key,
		Status: string(status)}

	res := p.db.Table("idempotency_keys").Clauses(clause.OnConflict{DoNothing: true}).Create(entity)
	if res.RowsAffected > 0 {
		s := statuses.Processing
		log.Printf("Added idempotency key %s. Status %s", key, s)

		return &s, nil
	} else if res.Error != nil {
		return nil, res.Error
	}

	res = p.db.First(entity)

	if entity.Status != string(statuses.Error) {
		return nil, defineStatusError(*entity)
	}

	res = p.db.Model(entity).Update("status", string(statuses.Processing)).Where("status = ?", string(statuses.Error))
	if res.RowsAffected == 0 {
		return nil, defineStatusError(*entity)
	}

	return (*statuses.IdempotencyStatus)(&entity.Status), res.Error
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

func defineStatusError(e domain.IdempotencyKey) error {
	return application.ProcessStatusRestrictions{Status: (statuses.IdempotencyStatus)(e.Status)}
}
