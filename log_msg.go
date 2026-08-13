package logger

import (
	"crypto/md5"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/google/uuid"
)
// Структура лога 
type LogMsg struct {
	Module string
	Level  string
	Code   int64
	Msg    string
	Debug  string
}

// Базовый конструктор
func NewLogMsg(module, level, msg string, code int64, debug string) *LogMsg {
	return &LogMsg{
		Module: module,
		Level:  level,
		Code:   code,
		Msg:    msg,
		Debug:  debug,
	}
}

/* Функция формирования hash для лога (pk), если указанно вставить или обновить формируется hash по данным, 
   если вставка в любом случае то формируется случайный hash
*/ 
func (l *LogMsg) Hash(upsert bool) uuid.UUID {
	if upsert {
		input := fmt.Sprintf("%s|%s|%d|%s", l.Module, l.Level, l.Code, l.Msg)
		hash := md5.Sum([]byte(input))
		return uuid.UUID(hash)
	}
	return uuid.New()
}

// Формирование всего лога в читабельный вид в консоли
func (l *LogMsg) String() string {
	return fmt.Sprintf("%s [%s] <%s>: %s (%d) >> %s", time.Now().Format("2006-01-02 15:04:05"), l.Level, l.Module, l.Msg, l.Code, l.Debug)
}
// Вспомогательная функция для определения bool статуса лога
func (l *LogMsg) Bool() bool {
	falseLevel := []string{"ERROR", "FATAL", "CRITICAL"}
	if slices.Contains(falseLevel, l.Level) {
		return false
	}
	return true
}
// Вызов падения приложения с выводом лога
func (l *LogMsg) Fatal() {
	fmt.Printf("%s\n", l.String())
	os.Exit(1)
}
// Вывод на консоль с фильтром. Если выключен выводятся только ошибки и критика
func (l *LogMsg) Print(all_flg bool) *LogMsg {
	if !all_flg {
		falseLevel := []string{"INFO", "WARNING"}
		if slices.Contains(falseLevel, l.Level) {
			return l
		}
	}
	fmt.Printf("%s\n", l.String())
	return l
}
//  Вспомогательная функция Если критика вызывает падение, если нет возвращает этот-же объект
func (l *LogMsg) Check() *LogMsg {
	falseLevel := []string{"FATAL", "CRITICAL"}
	if slices.Contains(falseLevel, l.Level) {
		l.Fatal()
	}
	return l
}