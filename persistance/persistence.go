package persistance

import (
	"errors"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	reconnectTime = 3 * time.Second
)

type DBQuerier struct {
	db *gorm.DB
}

func NewDBQuerier() (*DBQuerier) {
	for {
		db, err := gorm.Open(postgres.New(postgres.Config{
			DSN:                  "host=sws-postgres user=postgres password=pwadmin dbname=swsdb port=5432",
			PreferSimpleProtocol: true,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Printf("Failed to connect to DB: %v", err)
			time.Sleep(reconnectTime)
			continue
		}
		db.AutoMigrate(&Info{})
		return &DBQuerier{
			db: db,
		}
	}
}

type Info struct {
	ID uint `gorm:"primaryKey"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Transactions int
	Amount       float64
}

func (q *DBQuerier) Put(blockNum, transactionCount int, amount float64) {
	i := &Info{
		ID:           uint(blockNum),
		Transactions: transactionCount,
		Amount:       amount,
	}
	res := q.db.First(&i, blockNum)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			q.db.Create(i)
		}
		return
	}
	q.db.Save(&i)
}

func (q *DBQuerier) Get(blockNum int) (*Info, bool) {
	i := &Info{}
	res := q.db.First(&i, blockNum)
	if res.Error != nil {
		return nil, false
	}
	return i, true
}

func (q *DBQuerier) Shutdown() error {
	db, err := q.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
