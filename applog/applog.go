// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//
// Application logging functions
package applog

import (
	"fmt"
	"log"
	"strings"
)

func Error(msg string, args ... interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Printf("ERROR: " + msg, args...)
}

func Errorf(msg string, args ... interface{}) {
	if len(args) == 0 {
		fmt.Println("ERROR: ", msg)
	} else {
		msg += "\n"
		fmt.Printf("ERROR: " + msg, args...)
	}
}

func Info(msg string, args ... interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Printf("INFO: " + msg, args...)
}

func Infof(msg string, args ... interface{}) {
	if len(args) == 0 {
		fmt.Println("INFO: ", msg)
	} else {
		msg += "\n"
		fmt.Printf("INFO: " + msg, args...)
	}
}

func Fatal(msg string, args ... interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	log.Fatalf("FATAL: " + msg, args...)
}

func Fatalf(msg string, args ... interface{}) {
	if len(args) == 0 {
		log.Fatal("FATAL: " + msg)
	} else {
		msg += "\n"
		log.Fatalf("FATAL: " + msg, args...)
	}
}