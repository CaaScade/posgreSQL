package tail

import (
	"github.com/hpcloud/tail"
)

var buf chan string

func init() {
	buf = make(chan string, 1500)
}

func InitTail(file string) error {
	t, err := tail.TailFile(file, tail.Config{
		Follow: true,
		ReOpen: true,
	})
	if err != nil {
		return err
	}
	go func() {
		for line := range t.Lines {
			buf <- line.Text
		}
	}()
	return nil
}

func Tail() <-chan string {
	lines := make(chan string, 1500)
	go func() {
		for line := range buf {
			lines <- line
		}
	}()
	return lines
}
