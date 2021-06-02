package main

import (
	"errors"
	"time"

	"github.com/dude333/rapina/progress"
)

func main() {

	progress.Cursor(false)
	defer progress.Cursor(true)
	progress.Status("a status msg")

	progress.Running("start process")
	progress.Error(errors.New("some error"))
	time.Sleep(time.Second)
	progress.RunOK()

	progress.Running("start another process")
	time.Sleep(time.Second)
	progress.Status("middle")
	time.Sleep(time.Second)
	progress.RunFail()

	f1()

	progress.Running("start spinner")
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		progress.Spinner()
		if i == 20 {
			progress.Status("spinner interrupt")
		}
	}
	progress.RunOK()

	progress.Status("end.")
}

func f1() {
	progress.Running("Running *f1*")
	time.Sleep(time.Second)
	progress.Warning("f1 warning")
	time.Sleep(time.Second)
	progress.RunOK()
}
