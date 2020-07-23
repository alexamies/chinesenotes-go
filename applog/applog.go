/*
Application logging functions
*/
package applog

import (
	"log"
)

func Error(msg string, args ... interface{}) {
	log.Printf("ERROR: " + msg, args...)
}

func Errorf(msg string, args ... interface{}) {
	if len(args) == 0 {
		log.Println("ERROR: ", msg)
	} else {

		log.Printf("ERROR: " + msg, args...)
	}
}

func Info(msg string, args ... interface{}) {
	log.Printf("INFO: " + msg, args...)
}

func Infof(msg string, args ... interface{}) {
	if len(args) == 0 {
		log.Println("INFO: ", msg)
	} else {
		log.Printf("INFO: " + msg, args...)
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