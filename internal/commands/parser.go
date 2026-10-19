package commands

import (
	"bufio"
	"fmt"
	"io"
)

const (
	maxCommandArgs  = 128
	maxBulkLength   = 1 << 20 // 1 MiB per argument
	maxCommandBytes = 4 << 20 // 4 MiB per command
	maxHeaderLength = 64
)

// "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"

func ReadHeader(reader *bufio.Reader) (string, error) {
	header := make([]byte, 0, maxHeaderLength)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF && len(header) > 0 {
				return "", io.ErrUnexpectedEOF
			}

			return "", err
		}

		header = append(header, b)

		if len(header) > maxHeaderLength {
			return "", fmt.Errorf("RESP header too long")
		}

		if b == '\n' {
			if len(header) < 2 || header[len(header)-2] != '\r' {
				return "", fmt.Errorf("header must end with CRLF")
			}
			return string(header[:len(header)-2]), nil
		}
	}
}
