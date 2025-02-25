package internal

import (
	"github.com/busy-cloud/boat/pool"
	"github.com/go-co-op/gocron/v2"
	"time"
)

var scheduler gocron.Scheduler

func init() {
	var err error
	scheduler, err = gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

}

func Interval(interval int64, fn func()) (gocron.Job, error) {
	return scheduler.NewJob(
		gocron.DurationJob(time.Second*time.Duration(interval)),
		gocron.NewTask(func() { _ = pool.Insert(fn) }),
	)
}

func Crontab(crontab string, fn func()) (gocron.Job, error) {
	return scheduler.NewJob(
		gocron.CronJob(crontab, false),
		gocron.NewTask(func() { _ = pool.Insert(fn) }),
	)
}
