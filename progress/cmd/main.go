package main

import (
	"errors"
	"os"
	"time"

	"github.com/dude333/rapina/progress"
)

func main() {

	p := progress.Open(os.Stdout)
	defer p.Close()
	p.Status("a status msg")

	p.Running("start process")
	p.Error(errors.New("some error"))
	time.Sleep(time.Second)
	p.RunOK()

	p.Running("start another process")
	time.Sleep(time.Second)
	p.Status("middle")
	time.Sleep(time.Second)
	p.RunFail()

	p.Running("start spinner")
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		p.Spinner()
		if i == 20 {
			p.Status("spinner interrupt")
		}
	}
	p.RunOK()

	p.Status("end.")
}
