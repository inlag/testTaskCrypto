package database

import (
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func InitDb(url string) (*sqlx.DB, error) {
	db, errConnect := sqlx.Connect("pgx", url)
	if errConnect != nil {
		return nil, errors.Wrap(errConnect, "connect to database is failed")
	}

	if errPing := db.Ping(); errPing != nil {
		return nil, errors.Wrap(errPing, "ping to database is failed")
	}

	return db, nil
}

// AveragePrice модель среднего ценника
type AveragePrice struct {
	// start Указывает начало временного промежутка для среднего
	Start time.Time `json:"start" db:"start"`
	// end Указывает конец временного промежутка для среднего
	End time.Time `json:"end" db:"end"`
	// price Среднее значение
	Price float64 `json:"price" db:"amount"`
}

//  ReceivedMsg json возвращаемый от сервера blockchain.com
type ReceivedMsg struct {
	Price float64 `json:"price"` // [Timestamp Open High Low Close Volume]
}
