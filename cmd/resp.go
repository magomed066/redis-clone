package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func readArrayLength(reader *bufio.Reader) (int, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	if !strings.HasSuffix(line, "\r\n") {
		return 0, fmt.Errorf("array header must end with CRLF")
	}

	if !strings.HasPrefix(line, "*") {
		return 0, fmt.Errorf("array header must start with *")
	}

	digits := line[1 : len(line)-2]

	count, err := strconv.Atoi(digits)
	if err != nil {
		return 0, fmt.Errorf("invalid array length: %w", err)
	}

	if count <= 0 || count > 128 {
		return 0, fmt.Errorf("array length must be between 1 and 128")
	}

	return count, nil
}

func readBulkString(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(line, "$") {
		return "", fmt.Errorf("bulk string header must start with $")
	}

	if !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("bulk string header must end with CRLF")
	}

	digits := line[1 : len(line)-2]

	length, err := strconv.Atoi(digits)
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length: %w", err)
	}

	if length < 0 || length > 1<<20 {
		return "", fmt.Errorf("bulk string length must be between 0 and 1048576")
	}

	payload := make([]byte, length+2)

	if _, err := io.ReadFull(reader, payload); err != nil {
		return "", err
	}

	if payload[length] != '\r' || payload[length+1] != '\n' {
		return "", fmt.Errorf("bulk string must end with CRLF")
	}

	return string(payload[:length]), nil
}

func readCommand(reader *bufio.Reader) ([]string, error) {
	argCount, err := readArrayLength(reader)
	if err != nil {
		return []string{}, err
	}

	payload := make([]string, 0, argCount)

	for i := 0; i < argCount; i++ {
		val, err := readBulkString(reader)
		if err != nil {
			return []string{}, err
		}

		payload = append(payload, val)
	}

	return payload, nil
}

func encodeBulkString(value string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
}

func encodeInteger(value int64) string {
	return fmt.Sprintf(":%d\r\n", value)
}
