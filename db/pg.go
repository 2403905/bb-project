package db

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
)

type PgDatabase struct {
	db    *pg.DB
	dsn   string
	debug bool
}

// Hook for logging queries
type dbLogger struct{}

func (d dbLogger) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (d dbLogger) AfterQuery(ctx context.Context, qe *pg.QueryEvent) error {
	q, _ := qe.FormattedQuery()
	str := fmt.Sprintf("executing a query: %s", q)
	if qe.Err != nil {
		str = qe.Err.Error() + " " + str
	}
	fmt.Println(str)
	return nil
}

func InitConnection(dsn string, debug bool) (*PgDatabase, error) {
	conInstance := &PgDatabase{dsn: dsn, debug: debug}
	if err := conInstance.connect(); err != nil {
		return nil, err
	}
	return conInstance, nil
}

func (d *PgDatabase) GetDbE() (*pg.DB, error) {
	if d != nil {
		err := d.checkConnectionE()
		if err != nil {
			return nil, err
		}
	}
	return d.db, nil
}

func (d *PgDatabase) connect() error {
	opts, err := pg.ParseURL(d.dsn)
	if err != nil {
		return err
	}

	d.db = pg.Connect(opts)

	if d.debug {
		d.db.AddQueryHook(dbLogger{})
	}

	return nil
}

func (d *PgDatabase) checkConnectionE() error {
	var n int
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := d.db.QueryOneContext(ctx, pg.Scan(&n), "SELECT 1")
	if err == nil {
		return nil
	}
	fmt.Println("Connection to Postgres lost. Trying to reconnect", err)

	d.Close()

	for i := 0; i < 3; i++ {
		if err = d.connect(); err == nil {
			return nil
		}
		fmt.Println("Failed connect to Postgres. Trying to reconnect", err)
		time.Sleep(time.Second)
	}
	return err
}

func (d *PgDatabase) Close() {
	err := d.db.Close()
	if err != nil {
		fmt.Println(err)
	}
}
