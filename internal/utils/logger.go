package utils

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Log struct {
	namespace string
	name      string
	source    string
}

func NewLog(source string, namespace string, name string) *Log {
	l := new(Log)
	l.namespace = namespace
	l.name = name
	l.source = source
	return l
}

func (l *Log) sprintf(template string, args ...any) string {
	msg := fmt.Sprintf(template, args...)
	return fmt.Sprintf("[%s/%s]   %s", l.namespace, l.name, msg)
}

func (l *Log) Info(template string, args ...any) {
	log.Log.WithName(l.source).Info(l.sprintf(template, args...))
}

func (l *Log) Error(err error, template string, args ...any) {
	log.Log.WithName(l.source).Error(err, l.sprintf(template, args...))
}
