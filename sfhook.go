package sfhook

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

//WriterMap : Alternatively map a log level to an io.Writer
type WriterMap map[logrus.Level]Writer

//SFHook this is a hook for logrus
type SFHook struct {
	writerMap WriterMap
	levels    []logrus.Level
	formatter logrus.Formatter
	lock      *sync.Mutex
}

// NewHook : Given a WriterMap with keys equal to log levels.
func NewHook(levelMap map[logrus.Level]Writer, formatter logrus.Formatter) *SFHook {
	if formatter == nil {
		formatter = &logrus.TextFormatter{DisableColors: true}
	}

	hook := &SFHook{
		formatter: formatter,
		writerMap: levelMap,
		lock:      new(sync.Mutex),
	}

	for level := range levelMap {
		hook.levels = append(hook.levels, level)
	}

	return hook
}

//Fire : output log
func (s *SFHook) Fire(entry *logrus.Entry) error {
	var (
		msg string
		err error
		ok  bool
	)

	s.lock.Lock()

	if _, ok = s.writerMap[entry.Level]; !ok {
		err = fmt.Errorf("no writer provided for loglevel: %d", entry.Level)
		log.Println(err.Error())
		return err
	}
	entry.WithFields(logrus.Fields{
		"time":  entry.Time,
		"level": entry.Level,
	})
	msg, err = entry.String()

	if err != nil {
		log.Println("failed to generate string for entry:", err)
		return err
	}
	_, err = s.writerMap[entry.Level].Write([]byte(msg), entry.Level)
	s.lock.Unlock()
	return err
}

//Levels : Return configured log levels
func (s *SFHook) Levels() []logrus.Level {
	return s.levels
}

//Writer : interface for Hook's writer
type Writer interface {
	Write([]byte, logrus.Level) (int, error)
}

//SFWriter FileWriter for SFHook
type SFWriter struct {
	path             string
	unit             int
	fileNameFormater func(time.Time, logrus.Level) string
	writer           *bufio.Writer
	fileFirstWirte   time.Time //first write to file
	f                *os.File
}

//NewWriter : New SFWriter
//Params : path : directory of logFile
//         unit : unit of cutting logFile
//         fileNameFormater : callback for getting the name of logFile
func NewWriter(path string, unit int, fileNameFormater func(time.Time, logrus.Level) string) (*SFWriter, error) {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return nil, err
	}

	if unit <= 0 {
		unit = 1
	}
	if fileNameFormater == nil {
		fileNameFormater = func(t time.Time, level logrus.Level) string {
			return t.Format("2006-01-02") + ".log"
		}
	}
	return &SFWriter{
		path:             path,
		unit:             unit,
		fileNameFormater: fileNameFormater,
	}, nil
}

func (s *SFWriter) Write(p []byte, level logrus.Level) (int, error) {
	w, err := s.getWriter(level)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(p)
	if err == nil {
		w.Flush()
	}
	return n, err
}

// Write a log line directly to a file
func (s *SFWriter) getWriter(level logrus.Level) (*bufio.Writer, error) {
	t := time.Now()
	if t.Format("2006-01-02") == s.fileFirstWirte.Format("2006-01-02") {
		return s.writer, nil
	}
	s.writer = nil
	s.f.Close()
	var (
		path string
		err  error
	)
	path = s.path + "/" + s.fileNameFormater(t, level)

	s.f, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println("failed to open logfile:", path, err)
		return nil, err
	}

	s.writer = bufio.NewWriter(s.f)
	return s.writer, nil
}

// Close : close file
func (s *SFWriter) Close() {
	s.f.Close()
}
