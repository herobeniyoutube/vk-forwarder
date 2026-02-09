package domain

type IdempotencyKey struct {
	Key    string `gorm:"column:key;primaryKey"`
	Status string
}
