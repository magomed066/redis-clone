package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/magomed066/redis-clone/internal/store"
)

func executeCommand(db *store.Store, args []string) string {
	if len(args) == 0 {
		return "-ERR empty command\r\n"
	}

	var response string
	command := strings.ToUpper(args[0])

	switch command {
	case "PING":
		if len(args) != 1 {
			response = "-ERR PING arguments not supported yet\r\n"
		} else {
			response = "+PONG\r\n"
		}

	case "ECHO":
		if len(args) != 2 {
			response = "-ERR wrong number of arguments for ECHO\r\n"
		} else {
			response = encodeBulkString(args[1])
		}

	case "SET":
		if len(args) == 3 {
			db.Set(args[1], args[2])
			response = "+OK\r\n"
			break
		}

		if len(args) != 5 {
			response = "-ERR wrong number of arguments for SET\r\n"
			break
		}

		if strings.ToUpper(args[3]) != "PX" {
			response = "-ERR unsupported SET option\r\n"
			break
		}

		milliseconds, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil {
			response = "-ERR expiration must be a valid integer\r\n"
			break
		}

		maxMilliseconds := int64(math.MaxInt64) / int64(time.Millisecond)
		if milliseconds <= 0 || milliseconds > maxMilliseconds {
			response = "-ERR invalid expiration time for SET\r\n"
			break
		}

		ttl := time.Duration(milliseconds) * time.Millisecond

		err = db.SetWithTTL(args[1], args[2], ttl)
		if err != nil {
			response = "-ERR " + err.Error() + "\r\n"
			break
		}

		response = "+OK\r\n"

	case "GET":
		if len(args) != 2 {
			response = "-ERR wrong number of arguments for GET\r\n"
		} else {
			val, exists := db.Get(args[1])
			if !exists {
				response = "$-1\r\n"
			} else {
				response = encodeBulkString(val)
			}
		}

	case "INCR":
		if len(args) != 2 {
			response = "-ERR wrong number of arguments for INCR\r\n"
		} else {
			val, err := db.Incr(args[1])
			if err != nil {
				response = "-ERR " + err.Error() + "\r\n"
			} else {
				response = encodeInteger(val)
			}
		}

	case "DEL":
		if len(args) < 2 {
			response = "-ERR wrong number of arguments for DEL\r\n"
		} else {
			deleted := db.DeleteMany(args[1:])
			if deleted < 1 {
				response = encodeInteger(0)
			} else {
				response = encodeInteger(deleted)
			}
		}

	case "EXPIRE":
		if len(args) != 3 {
			response = "-ERR wrong number of arguments for EXPIRE\r\n"
			break
		}

		seconds, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			response = "-ERR seconds must be a valid integer\r\n"
			break
		}

		// Check before multiplying to avoid overflowing time.Duration.
		maxSeconds := int64(math.MaxInt64) / int64(time.Second)
		if seconds > maxSeconds {
			response = "-ERR expiration duration is too large\r\n"
			break
		}

		// Zero and negative seconds mean immediate deletion.
		var ttl time.Duration
		if seconds > 0 {
			ttl = time.Duration(seconds) * time.Second
		}

		if db.Expire(args[1], ttl) {
			response = encodeInteger(1)
		} else {
			response = encodeInteger(0)
		}

	case "TTL":
		if len(args) != 2 {
			response = "-ERR wrong number of arguments for TTL\r\n"
		} else {
			response = encodeInteger(db.TTL(args[1]))
		}

	default:
		response = "-ERR unknown command\r\n"
	}

	return response
}
