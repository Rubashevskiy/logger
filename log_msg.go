package logger

import (
	"crypto/md5"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/google/uuid"
)

type LogMsg struct {
	Module string `db:"module"`
	Level  string `db:"level"`
	Code   int64  `db:"code"`
	Msg    string `db:"msg"`
	Debug  string `db:"debug"`
}

func NewLogMsg(module, level, msg string, code int64, debug string) *LogMsg {
	return &LogMsg{
		Module: module,
		Level:  level,
		Code:   code,
		Msg:    msg,
		Debug:  debug,
	}
}

func (l *LogMsg) Hash(upsert bool) uuid.UUID {
	if upsert {
		input := fmt.Sprintf("%s|%s|%d|%s", l.Module, l.Level, l.Code, l.Msg)
		hash := md5.Sum([]byte(input))
		return uuid.UUID(hash)
	}
	return uuid.New()
}

func (l *LogMsg) String() string {
	return fmt.Sprintf("%s [%s] <%s>: %s (%d) >> %s", time.Now().Format("2006-01-02 15:04:05"), l.Level, l.Module, l.Msg, l.Code, l.Debug)
}

func (l *LogMsg) Bool() bool {
	falseLevel := []string{"ERROR", "FATAL", "CRITICAL"}
	if slices.Contains(falseLevel, l.Level) { // Убрали string()
		return false
	}
	return true
}

func (l *LogMsg) Fatal() {
	fmt.Printf("%s\n", l.String())
	os.Exit(1)
}

func (l *LogMsg) Print(all_flg bool) *LogMsg {
	if !all_flg {
		falseLevel := []string{"INFO", "WARNING"}
		if slices.Contains(falseLevel, l.Level) { // Убрали string()
			return l
		}
	}
	fmt.Printf("%s\n", l.String())
	return l
}

func (l *LogMsg) Check() *LogMsg {
	falseLevel := []string{"FATAL", "CRITICAL"}
	if slices.Contains(falseLevel, l.Level) { // Убрали string()
		l.Fatal()
	}
	return l
}