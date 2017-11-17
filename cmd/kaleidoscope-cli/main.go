package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/monochromegane/kaleidoscope"
)

func main() {
	kes, err := kaleidoscope.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		r := csv.NewReader(bytes.NewReader(scanner.Bytes()))
		r.Comma = ' '
		record, err := r.Read()
		if err != nil && err != io.EOF {
			fmt.Print(err.Error())
			continue
		}
		commands := compact(record)
		if len(commands) == 0 {
			continue
		}
		out, err := run(&kes, commands)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if out != "" {
			fmt.Println(out)
		}
	}
}

func run(kes *kaleidoscope.Kaleidoscope, commands []string) (string, error) {
	switch strings.ToLower(commands[0]) {
	case "create":
		dbname := commands[1]
		size := 2048
		if len(commands) > 2 {
			s := commands[2]
			is, err := strconv.Atoi(s)
			if err != nil {
				return "", err
			}
			size = is
		}
		return kes.Create(dbname, size)
	case "save":
		return "", kes.Save()
	case "use":
		return "", kes.Use(commands[1])
	case "set":
		key := commands[1]
		value := commands[2]
		return kes.Set(key, value)
	case "get":
		_, out, err := kes.Get(commands[1])
		return string(out), err
	case "del":
		return kes.Del(commands[1])
	case "sync":
		return "", kes.StartSync()
	default:
		return "", fmt.Errorf("Unknown command: %s", commands[0])
	}
}

func compact(src []string) []string {
	var result []string
	for _, s := range src {
		if s == "" {
			continue
		}
		result = append(result, s)
	}
	return result
}
