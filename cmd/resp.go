package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func readHeader(reader *bufio.Reader) (string, error) {
	const maxHeaderBytes = 64
	header := make([]byte, 0, maxHeaderBytes)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF && len(header) > 0 {
				return "", io.ErrUnexpectedEOF
			}
			return "", err
		}

		header = append(header, b)

		if b == '\n' {
			if len(header) < 2 || header[len(header)-2] != '\r' {
				return "", fmt.Errorf("header must end with CRLF")
			}

			return string(header), nil
		}

		if len(header) >= maxHeaderBytes {
			return "", fmt.Errorf("header exceeds %d bytes", maxHeaderBytes)
		}
	}
}

func readArrayLength(reader *bufio.Reader) (int, error) {
	line, err := readHeader(reader)
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

func readBulkString(reader *bufio.Reader, remainingBytes int) (string, error) {
	line, err := readHeader(reader)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(line, "$") {
		return "", fmt.Errorf("bulk string header must start with $")
	}

	digits := line[1 : len(line)-2]

	length, err := strconv.Atoi(digits)
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length: %w", err)
	}

	const maxBulkBytes = 1 << 20 // 1 MiB

	if length < 0 || length > maxBulkBytes {
		return "", fmt.Errorf(
			"bulk string length must be between 0 and %d",
			maxBulkBytes,
		)
	}

	// Check the command budget before allocating the payload.
	if length > remainingBytes {
		return "", fmt.Errorf("command payload exceeds size limit")
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
		return nil, err
	}

	const maxCommandBytes = 4 << 20 // 4 MiB
	remainingBytes := maxCommandBytes

	args := make([]string, 0, argCount)

	for i := 0; i < argCount; i++ {
		value, err := readBulkString(reader, remainingBytes)
		if err != nil {
			// The array header was already read, so EOF here
			// means the command is incomplete.
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}

		remainingBytes -= len(value)
		args = append(args, value)
	}

	return args, nil
}

func encodeBulkString(value string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
}

func encodeInteger(value int64) string {
	return fmt.Sprintf(":%d\r\n", value)
}
