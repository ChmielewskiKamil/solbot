package parser

import (
	"fmt"
	"strings"
)

// Tracing code taken from Thorsten Ball's book "Writing An Interpreter In Go"

var traceLevel int = 0

const traceIdentPlaceholder string = "\t"

func identLevel() string {
	return strings.Repeat(traceIdentPlaceholder, traceLevel-1)
}

func tracePrint(msg string) {
	fmt.Printf("%s%s\n", identLevel(), msg)
}

func incIdent() { traceLevel = traceLevel + 1 }
func decIdent() { traceLevel = traceLevel - 1 }

func trace(msg string) string {
	incIdent()
	tracePrint("BEGIN " + msg)
	return msg
}

func un(msg string) {
	tracePrint("END " + msg)
	decIdent()
}
