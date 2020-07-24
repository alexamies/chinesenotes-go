/*
Application logging functions
*/
package applog

import (
	"fmt"
	"log"
)

func Error(msg string, args ... interface{}) {
	fmt.Printf("ERROR: " + msg, args...)
}

func Errorf(msg string, args ... interface{}) {
	if len(args) == 0 {
		fmt.Println("ERROR: ", msg)
	} else {

		fmt.Printf("ERROR: " + msg, args...)
	}
}

func Info(msg string, args ... interface{}) {
	fmt.Printf("INFO: " + msg, args...)
}

func Infof(msg string, args ... interface{}) {
	if len(args) == 0 {
		fmt.Println("INFO: ", msg)
	} else {
		fmt.Printf("INFO: " + msg, args...)
	}
}

func Fatal(msg string, args ... interface{}) {
	log.Fatalf("FATAL: " + msg, args...)
}

func Fatalf(msg string, args ... interface{}) {
	if len(args) == 0 {
		log.Fatal("FATAL: " + msg)
	} else {
		log.Fatalf("FATAL: " + msg, args...)
	}
}