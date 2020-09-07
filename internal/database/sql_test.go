package database

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/inlag/testTaskCrypto/internal/config"

	"github.com/jmoiron/sqlx"
)

var sqlPool *sqlx.DB

func init() {
	dbPool, errInitDB := InitDb(config.GetDBUrl())
	if errInitDB != nil {
		log.Fatalln(errInitDB)
	}
	sqlPool = dbPool
}

func TestSql_SetPool(t *testing.T) {
	cases := []struct {
		name    string
		pool    *sqlx.DB
		wantErr bool
	}{
		{
			name:    "TestWithEmptyPool",
			pool:    nil,
			wantErr: true,
		},
		{
			name:    "TestWithNotEmptyPool",
			pool:    &sqlx.DB{},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var s = &Sql{}
			errSetPool := s.SetPool(tt.pool)
			if (errSetPool != nil) != tt.wantErr {
				t.Errorf("test SetPool is failed, want %v receive %v", tt.wantErr, errSetPool)
			}
		})
	}
}

func TestSql_SaveAverage(t *testing.T) {
	var cases = []struct {
		name    string
		avg     AveragePrice
		wantErr bool
	}{
		{
			name:    "TestWithEmptyTime",
			avg:     AveragePrice{},
			wantErr: true,
		},
		{
			name: "TestWithEmptyPrice",
			avg: AveragePrice{
				Start: time.Now(),
				End:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "TestWithCorrectAverage",
			avg: AveragePrice{
				Start: time.Now(),
				End:   time.Now(),
				Price: 4555,
			},
			wantErr: false,
		},
	}

	var repository = Sql{
		pool: sqlPool,
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			errSaveAverage := repository.SaveAverage(tt.avg)
			if (errSaveAverage != nil) != tt.wantErr {
				t.Errorf("test SaveAverage is failed, want %v receive %v", tt.wantErr, errSaveAverage)
			}
		})
	}
}

func TestSql_GetAverage(t *testing.T) {
	var cases = []struct {
		name    string
		sql     Sql
		ch      chan []AveragePrice
		wantErr bool
	}{
		{
			name: "TestWithEmptyChannel",
			sql: Sql{
				pool: sqlPool,
			},
			wantErr: true,
		},
		{
			name: "TestWithCorrect",
			sql: Sql{
				pool: sqlPool,
			},
			ch:      make(chan []AveragePrice),
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				<-tt.ch
			}()
			errGetAverage := tt.sql.GetAverage(tt.ch)
			if (errGetAverage != nil) != tt.wantErr {
				t.Errorf("test GetAverage is failed, want %v receive %v", tt.wantErr, errGetAverage)
			}
		})
	}
}

func TestSql_CleanupAverage(t *testing.T) {
	var cases = []struct {
		name     string
		generate int
		wantErr  bool
	}{
		{
			name:     "TestWithoutRows",
			generate: 0,
			wantErr:  true,
		},
		{
			name:     "TestWithRows",
			generate: 10,
			wantErr:  false,
		},
	}

	var s = Sql{
		pool: sqlPool,
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			generateOldAverageData(tt.generate)
			errCleanup := s.CleanupAverage()
			if (errCleanup != nil) != tt.wantErr {
				t.Errorf("test CleanupAverage is failed, want: %v receive: %v", tt.wantErr, errCleanup)
			}
		})
	}
}

func generateOldAverageData(count int) {
	if count == 0 {
		return
	}
	var avgs []AveragePrice
	oldTime := time.Now().Add(time.Hour * -2)
	for i := 0; i <= count; i++ {
		avg := AveragePrice{
			Start: oldTime.Add(time.Minute * 5),
			End:   oldTime.Add(time.Minute * 5),
			Price: rand.Float64(),
		}
		avgs = append(avgs, avg)
	}

	for _, average := range avgs {
		_, _ = sqlPool.Exec(`INSERT INTO average_price("start", "end","amount")  VALUES ($1,$2,$3)`, average.Start, average.End, average.Price)
	}
}

func Test_saveTimeHandler(t *testing.T) {
	type args struct {
		startTime time.Time
		amount    float64
		i         float64
		errChan   chan error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestWithNullDataTime",
			args: args{
				startTime: time.Time{},
				amount:    0,
				i:         0,
				errChan:   make(chan error),
			},
			wantErr: true,
		},
		{
			name: "TestWithNullDataAmount",
			args: args{
				startTime: time.Now(),
				amount:    0,
				i:         0,
				errChan:   make(chan error),
			},
			wantErr: true,
		},
		{
			name: "TestWithCorrectData",
			args: args{
				startTime: time.Now(),
				amount:    123,
				i:         2,
				errChan:   make(chan error),
			},
			wantErr: false,
		},
	}

	var s = &Sql{
		pool: sqlPool,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				for {
					select {
					case errSaveHandler := <-tt.args.errChan:
						if (errSaveHandler != nil) != tt.wantErr {
							t.Errorf("test CleanupAverage is failed, want: %v receive: %v", tt.wantErr, errSaveHandler)
						}
					default:
					}
				}
			}()
			saveTimeHandler(tt.args.startTime, tt.args.amount, tt.args.i, s, tt.args.errChan)
		})
	}
}
