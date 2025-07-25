// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 🤖 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math"
	"net"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

const (
	toLowerTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@abcdefghijklmnopqrstuvwxyz[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
	toUpperTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`ABCDEFGHIJKLMNOPQRSTUVWXYZ{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
)

// Copyright © 2014, Roger Peppe
// github.com/rogpeppe/fastuuid
// All rights reserved.

var (
	uuidSeed    [24]byte
	uuidCounter uint64
	uuidSetup   sync.Once
)

// UUID generates an universally unique identifier (UUID)
func UUID() string {
	// Setup seed & counter once
	uuidSetup.Do(func() {
		if _, err := rand.Read(uuidSeed[:]); err != nil {
			return
		}
		uuidCounter = binary.LittleEndian.Uint64(uuidSeed[:8])
	})
	if atomic.LoadUint64(&uuidCounter) <= 0 {
		return "00000000-0000-0000-0000-000000000000"
	}
	// first 8 bytes differ, taking a slice of the first 16 bytes
	x := atomic.AddUint64(&uuidCounter, 1)
	_uuid := uuidSeed
	binary.LittleEndian.PutUint64(_uuid[:8], x)
	_uuid[6], _uuid[9] = _uuid[9], _uuid[6]

	// RFC4122 v4
	_uuid[6] = (_uuid[6] & 0x0f) | 0x40
	_uuid[8] = _uuid[8]&0x3f | 0x80

	// create UUID representation of the first 128 bits
	b := make([]byte, 36)
	hex.Encode(b[0:8], _uuid[0:4])
	b[8] = '-'
	hex.Encode(b[9:13], _uuid[4:6])
	b[13] = '-'
	hex.Encode(b[14:18], _uuid[6:8])
	b[18] = '-'
	hex.Encode(b[19:23], _uuid[8:10])
	b[23] = '-'
	hex.Encode(b[24:], _uuid[10:16])

	return UnsafeString(b)
}

// UUIDv4 returns a Random (Version 4) UUID.
// The strength of the UUIDs is based on the strength of the crypto/rand package.
func UUIDv4() string {
	token, err := uuid.NewRandom()
	if err != nil {
		return UUID()
	}
	return token.String()
}

// FunctionName returns function name
func FunctionName(fn any) string {
	if fn == nil {
		return ""
	}
	v := reflect.ValueOf(fn)
	if v.Kind() == reflect.Func {
		if v.IsNil() {
			return ""
		}
		pc := v.Pointer()
		f := runtime.FuncForPC(pc)
		if f == nil {
			return ""
		}
		return f.Name()
	}
	return v.Type().String()
}

// GetArgument check if key is in arguments
func GetArgument(arg string) bool {
	for i := range os.Args[1:] {
		if os.Args[1:][i] == arg {
			return true
		}
	}
	return false
}

// IncrementIPRange Find available next IP address
func IncrementIPRange(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ConvertToBytes returns integer size of bytes from human-readable string, ex. 42kb, 42M
// Returns 0 if string is unrecognized
func ConvertToBytes(humanReadableString string) int {
	strLen := len(humanReadableString)
	if strLen == 0 {
		return 0
	}

	var unitPrefixPos, lastNumberPos int
	// loop backwards to find the last numeric character and the unit prefix
	for i := strLen - 1; i >= 0; i-- {
		c := humanReadableString[i]
		if c >= '0' && c <= '9' {
			lastNumberPos = i
			break
		}
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			unitPrefixPos = i
		}
	}

	numPart := humanReadableString[:lastNumberPos+1]
	var size float64
	if strings.IndexByte(numPart, '.') >= 0 {
		var err error
		size, err = strconv.ParseFloat(numPart, 64)
		if err != nil {
			return 0
		}
	} else {
		i64, err := strconv.ParseUint(numPart, 10, 64)
		if err != nil {
			return 0
		}
		size = float64(i64)
	}

	if unitPrefixPos > 0 {
		switch humanReadableString[unitPrefixPos] {
		case 'k', 'K':
			size *= 1e3
		case 'm', 'M':
			size *= 1e6
		case 'g', 'G':
			size *= 1e9
		case 't', 'T':
			size *= 1e12
		case 'p', 'P':
			size *= 1e15
		}
	}

	if size > float64(math.MaxInt) {
		return math.MaxInt
	}

	return int(size)
}
