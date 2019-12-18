package sfhook

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const msg = "test message."
const msgNotOutput = "test2 message."

func TestLogEntryWritten(t *testing.T) {
	log := logrus.New()

	path := "logs"

	w, err := NewWriter(path, 1, nil)
	if err != nil {
		t.Errorf("mkdirAll [%s] error!", path)
	}
	hook := NewHook(WriterMap{
		logrus.InfoLevel: w,
	}, nil)
	log.Formatter = &logrus.JSONFormatter{}
	log.Hooks.Add(hook)

	log.Info(msg)
	log.Warn(msgNotOutput)

	fname := "logs/" + time.Now().Format("2006-01-02") + ".log"
	fi, err := os.Open(fname)
	if err != nil {
		t.Errorf("Open file [%s] error!", fname)
	} else {
		contents, err := ioutil.ReadAll(fi)
		if err != nil {
			t.Errorf("Error while reading from tmpfile: %s", err)
		}
		lines := strings.Split(string(contents), "\n")
		if len(lines) <= 0 {
			t.Errorf("Message read line number (%d) doesnt match written line number 1 for file: %s", len(lines), fname)
		} else if len(lines) > 1 || len(lines) == 2 && len(lines[1]) == 0 {
			t.Errorf("Message read line number (%d) doesnt match written line number 1 for file: %s", len(lines), fname)
		}
		if !bytes.Contains(contents, []byte("\"msg\":\""+msg+"\"")) {
			t.Errorf("Message read (%s) doesnt match message written (%s) for file: %s", contents, msg, fname)
		}

		if bytes.Contains(contents, []byte("\"msg\":\""+msgNotOutput+"\"")) {
			t.Errorf("Message read (%s) contains message written (%s) for file: %s", contents, msgNotOutput, fname)
		}
	}
	w.Close()
}

// func TestRuntime(t *testing.T) {
// 	log := logrus.New()

// 	path := "logs"

// 	w, err := NewWriter(path, 1, nil)
// 	if err != nil {
// 		t.Errorf("mkdirAll [%s] error!", path)
// 	}
// 	hook := NewHook(WriterMap{
// 		logrus.InfoLevel: w,
// 	}, nil)
// 	log.Formatter = &logrus.JSONFormatter{}
// 	log.Hooks.Add(hook)
// 	start := time.Now()
// 	for index := 0; index < 100000; index++ {
// 		funcName, file, line, ok := runtime.Caller(0)
// 		if ok {
// 			log.Info("Func Name=" + runtime.FuncForPC(funcName).Name())
// 			log.Info("Func Name=" + fmt.Sprintf("file: %s line=%d\n", file, line))
// 		}
// 	}
// 	log.Info(fmt.Sprintln("test1 run :", time.Now().Sub(start).Seconds(), "s"))
// }
