package database

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Sql struct {
	pool *sqlx.DB
}

func (s *Sql) SetPool(db *sqlx.DB) error {
	if db == nil {
		return errors.New("database is empty")
	}
	s.pool = db
	return nil
}

func (s *Sql) Calculate(fromWssChan chan float64, sqlToHttp chan []AveragePrice, errChan chan error, closeChan chan struct{}) {
	var (
		i                 float64
		amount            float64
		startTime         = time.Now()
		saveTimerDuration = time.Minute * 5

		cleanupTimer = time.NewTimer(time.Hour * 24)
		saveTimer    = time.NewTimer(saveTimerDuration)
	)

	log.Println("Открываем Sql соединение")
	for {
		select {
		case val := <-fromWssChan:
			i++
			amount += val
		case <-saveTimer.C:
			saveTimeHandler(startTime, amount, i, s, errChan)

			i, amount = 0, 0
			startTime = time.Now()
			saveTimer.Reset(saveTimerDuration)

		case <-cleanupTimer.C:
			if errCleanup := s.CleanupAverage(); errCleanup != nil {
				errChan <- errCleanup
			}
			cleanupTimer.Reset(time.Hour * 24)

		case <-sqlToHttp:
			errGetAvg := s.GetAverage(sqlToHttp)
			if errGetAvg != nil {
				errChan <- errGetAvg
				sqlToHttp <- []AveragePrice{}
			}
		case <-closeChan:
			saveTimeHandler(startTime, amount, i, s, errChan)

			closeChan <- struct{}{}
			return
		default:
		}
	}
}

func (s *Sql) SaveAverage(average AveragePrice) error {
	if average.Start.IsZero() || average.End.IsZero() {
		return errors.New("time is empty")
	}

	if average.Price == 0 {
		return errors.New("price is empty")
	}
	_, errExec := s.pool.Exec(`INSERT INTO average_price("start", "end","amount")  VALUES ($1,$2,$3)`, average.Start, average.End, average.Price)
	if errExec != nil {
		return errExec
	}
	return nil
}

func (s *Sql) GetAverage(a chan []AveragePrice) error {
	if a == nil {
		return errors.New("a send channel is empty")
	}

	res := []AveragePrice{}
	errQuery := s.pool.Select(&res, `SELECT "start", "end", "amount" FROM average_price`)
	if errQuery != nil {
		return errQuery
	}

	a <- res
	return nil
}

func (s *Sql) CleanupAverage() error {

	result, errExec := s.pool.Exec(`DELETE FROM average_price WHERE "end"<$1`, time.Now().AddDate(0, 0, -1))
	if errExec != nil {
		return errors.Wrap(errExec, "deleting rows is failed")
	}

	rows, errRowAff := result.RowsAffected()
	if errRowAff != nil {
		return errors.Wrap(errRowAff, "could not affected rows")
	}

	if rows == 0 {
		return errors.New("a query with deleting for the day is not affected rows")
	}

	return nil
}

func saveTimeHandler(startTime time.Time, amount, i float64, s *Sql, errChan chan error) {
	if amount == 0 && i == 0 {
		errChan <- errors.New("amount is empty")
	}
	avg := AveragePrice{
		Start: startTime,
		End:   time.Now(),
		Price: amount / i,
	}

	errSaveAvg := s.SaveAverage(avg)
	if errSaveAvg != nil {
		errChan <- errors.Wrap(errSaveAvg, "insert row is failed")
	}
}
