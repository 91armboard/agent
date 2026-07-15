package logger

/*
import (
	"log"
	"os"
)

var CLog *log.Logger

func init() {
	CLog = log.New(os.Stdout, "", 0)
}*/

import (
	"io"
	"log"
	"os"
)

type customLogger struct {
	logger *log.Logger
}

func NewCustomLogger(out io.Writer) *customLogger {
	logger := log.New(out, "", 0)
	return &customLogger{logger: logger}
}

func (c *customLogger) Println(v ...interface{}) {
	//c.logger.Printf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	c.logger.Println(v...)
}

func (c *customLogger) Printf(format string, v ...interface{}) {
	//c.logger.Printf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	c.logger.Printf(format, v...)
}

func (c *customLogger) Fatal(v ...interface{}) {
	//c.logger.Printf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	c.logger.Fatal(v...)
}

func (c *customLogger) Fatalln(v ...interface{}) {
	//c.logger.Printf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	c.logger.Fatalln(v...)
}

var Log = NewCustomLogger(os.Stdout)

// ////////////////////////////////////////////////////////////////////////////////
