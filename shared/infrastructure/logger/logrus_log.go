package logger

import (
	"context"
	"fmt"
	"infrastructure/shared/gogen"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func NewLogrusLog(appData gogen.ApplicationData) Logger {

	path := "."
	name := appData.AppName
	maxAge := 1

	writer, _ := rotatelogs.New(
		fmt.Sprintf("%s/logs/%s.log.%s", path, name, "%Y%m%d"),
		rotatelogs.WithLinkName(fmt.Sprintf("%s/%s.log", path, name)),
		rotatelogs.WithMaxAge(time.Duration(maxAge*24)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(1*24)*time.Hour),
	)

	theLogger := logrus.New()
	theLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "0102 150405.000",
	})

	theLogger.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.DebugLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		},
		theLogger.Formatter,
	))

	return &logrusLog{
		appData:   appData,
		theLogger: theLogger,
	}
}

type logrusLog struct {
	theLogger *logrus.Logger
	appData   gogen.ApplicationData
}

func (l logrusLog) Info(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.printLog(ctx, "INFO", messageWithArgs)
}

func (l logrusLog) Error(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.printLog(ctx, "ERROR", messageWithArgs)
}

func (l logrusLog) printLog(ctx context.Context, flag string, data any) {
	traceID := GetTraceID(ctx)

	msg := fmt.Sprintf("%-5s %s %-60v %s\n", flag, traceID, data, getFileLocationInfo(3))

	if flag == "INFO" {
		l.theLogger.Info(msg)
		return
	}

	if flag == "ERROR" {
		l.theLogger.Error(msg)
		return
	}

}
