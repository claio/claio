/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
