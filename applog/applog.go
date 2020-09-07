/*
Application logging functions
*/
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