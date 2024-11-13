package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"
)

type RDBParser = Parser

var (
	eofFlag = byte(0xFF) // End of the RDB file
	selFlag = byte(0xFE) // Database Selector
	exsFlag = byte(0xFD) // Expire time in seconds, see Key Expiry Timestamp
	exmFlag = byte(0xFC) // Expire time in milliseconds, see Key Expiry Timestamp
	htsFlag = byte(0xFB) // Hash table sizes for the main keyspace and expires, see Resizedb information
	auxFlag = byte(0xFA) // Auxiliary fields. Arbitrary key-value settings, see Auxiliary fields
)

type rdbParser struct {
	dbfile io.Reader
}

func NewRDBParser(dbfile io.Reader) RDBParser {
	return &rdbParser{dbfile}
}

func (r *rdbParser) Parse() ParseResponse {
	// Header section
	// Parse magic string & version -- 9 bytes
	headerBuf := make([]byte, 9)
	n, err := r.dbfile.Read(headerBuf)
	if err != nil {
		log.Println("[RDBParser] Error reading header: ", err.Error())
		return nil
	}
	if n != 9 {
		log.Println("[RDBParser] Incorrect number of header bytes read: ", n)
		return nil
	}

	// Assert magic string
	if ms := string(headerBuf[:5]); ms != "REDIS" {
		log.Println("[RDBParser] Improper magic string: ", ms)
		return nil
	}
	// Parse version
	version, err := strconv.Atoi(string(headerBuf[5:]))
	if err != nil {
		log.Println("[RDBParser] Error parsing version: ", err.Error())
		return nil
	}
	log.Println("[RDBParser] Version: ", version)

	// Parse sections
	// Extract OpCode from first bit
	oc, err := r.readSingleByte()
	if err != nil {
		log.Println("[RDBParser] Error reading first op code: ", err.Error())
		return nil
	}

	data, err := r.processOpCode(oc, [][]interface{}{})
	if err != nil && err != io.EOF {
		log.Println("[RDBParser] Error parsing file: ", err.Error())
		return nil
	}

	return data
}

// Recursively process opCode and subsequent data until EOF or error; returns
// list of k/v with optional expiry -- k, v[, e]
func (r *rdbParser) processOpCode(code byte, data [][]interface{}) ([][]interface{}, error) {
	switch code {
	case eofFlag:
		// We've reached the end of the file
		// Discard 8 byte checksum
		buf := make([]byte, 8)
		if _, err := r.dbfile.Read(buf); err != nil {
			return data, err
		}
		return data, nil
	case auxFlag:
		// Metadata; read two strings
		s1, err := r.readStringEncoding()
		if err != nil {
			return data, err
		}
		s2, err := r.readStringEncoding()
		if err != nil {
			return data, err
		}
		// Log metadata; we otherwise ignore it for now.
		log.Printf("[RDBParser] Metadata %q = %q\n", s1, s2)
		// Read next opCode
		oc, err := r.readSingleByte()
		if err != nil {
			return data, err
		}
		return r.processOpCode(oc, data)
	case selFlag:
		// Data section; read size encoded db index.
		idx, _, err := r.readSizeEncoding()
		if err != nil {
			return data, err
		}
		// Log idx; we otherwise ignore it for now.
		log.Printf("[RDBParser] Database index %d\n", idx)

		// Read next op code.
		oc, err := r.readSingleByte()
		if err != nil {
			return data, err
		}
		// Assert next op code is htsFlag.
		if oc != htsFlag {
			return data, fmt.Errorf("expected %#v op code, received %#v", htsFlag, oc)
		}
		// Read two size encoded integers.
		dbHashTableSize, _, err := r.readSizeEncoding()
		if err != nil {
			return data, err
		}
		expiryHashTableSize, _, err := r.readSizeEncoding()
		if err != nil {
			return data, err
		}
		// Log hash table sizes; otherwise ignore for now.
		log.Printf("[RDBParser] dbhts: %d; ehts: %d\n", dbHashTableSize, expiryHashTableSize)

		// Process data section
		nextOc, kvs, err := r.processValues([][]interface{}{})
		if err != nil {
			return data, err
		}
		return r.processOpCode(nextOc, append(data, kvs[:]...))
	default:
		return data, fmt.Errorf("unrecognized op code '%#v'", code)
	}
}

// Reads key/value pairs from database section with optional expiry.
// [["key", "val"], ["key", "val", "expiry"]]
func (r *rdbParser) processValues(data [][]interface{}) (byte, [][]interface{}, error) {
	// Initialize optional expiry
	var expiry time.Time
	var vt byte

	// Read first byte
	fb, err := r.readSingleByte()
	if err != nil {
		return byte(0), data, err
	}
	switch fb {
	case selFlag, eofFlag:
		// Finished reading this db; return op code & data.
		return fb, data, nil
	case exmFlag:
		// Read expiry time in ms; 8 byte unsigned long, little endian
		var exp int64
		if err := binary.Read(r.dbfile, binary.LittleEndian, &exp); err != nil {
			return byte(0), data, err
		}
		expiry = time.UnixMilli(exp)
		vt, err = r.readSingleByte()
		if err != nil {
			return byte(0), data, err
		}
	case exsFlag:
		// Read expiry time in s; 4 byte unsigned int, little endian
		var exp int32
		if err := binary.Read(r.dbfile, binary.LittleEndian, &exp); err != nil {
			return byte(0), data, err
		}
		expiry = time.Unix(int64(exp), 0)
		vt, err = r.readSingleByte()
		if err != nil {
			return byte(0), data, err
		}
	default: // No expiry, process k/v pair
		vt = fb
	}

	// Read string encoded key
	key, err := r.readStringEncoding()
	if err != nil {
		return byte(0), data, err
	}
	val, err := r.readValue(vt)
	if err != nil {
		return byte(0), data, err
	}
	// Return k/v with optional expiry
	if !expiry.IsZero() {
		return r.processValues(append(data, []interface{}{key, val, expiry}))
	}
	return r.processValues(append(data, []interface{}{key, val}))
}

// Reads a single byte.
func (r *rdbParser) readSingleByte() (byte, error) {
	buf := make([]byte, 1)
	if _, err := r.dbfile.Read(buf); err != nil {
		return byte(0), err
	}
	return buf[0], nil
}

// Reads size encoding from next byte.
// isString flag is set when the two significant bits of the next byte are 0b11.
// https://rdb.fnordig.de/file_format.html#integers-as-string
func (r *rdbParser) readSizeEncoding() (size int, isString bool, err error) {
	// Pull a bit off to get a size
	b, err := r.readSingleByte()
	if err != nil {
		return
	}
	// Switch on two most significant bits
	switch b >> 6 {
	case 0b11: // String formatting
		isString = true
		// Switch on the last six bits of the flag byte.
		switch b & 0b00111111 {
		case 0b0: // 8 bit integer.
			var val int8
			if err = binary.Read(r.dbfile, binary.LittleEndian, &val); err != nil {
				return
			}
			size = int(val)
		case 0b1: // 16 bit integer
			var val int16
			if err = binary.Read(r.dbfile, binary.LittleEndian, &val); err != nil {
				return
			}
			size = int(val)
		case 0b10: // 32 bit integer.
			var val int32
			if err = binary.Read(r.dbfile, binary.LittleEndian, &val); err != nil {
				return
			}
			size = int(val)
		default:
			err = fmt.Errorf("unrecognized integer string encoding: %#b", b&0b00111111)
			return
		}
	case 0b10: // Size is in next 4 bytes (32 bits)
		var val int32
		if err = binary.Read(r.dbfile, binary.LittleEndian, &val); err != nil {
			return
		}
		size = int(val)
	case 0b1: // Size is in remaining 6 bits plus next byte
		var nb byte
		nb, err = r.readSingleByte()
		if err != nil {
			return
		}
		sizeRaw := []byte{byte(b & 0b00111111), nb}
		reader := bytes.NewReader(sizeRaw)
		var val int16
		if err = binary.Read(reader, binary.LittleEndian, &val); err != nil {
			return
		}
		size = int(val)
	default:
		size = int(b)
	}
	return
}

func (r *rdbParser) readStringEncoding() ([]byte, error) {
	size, isString, err := r.readSizeEncoding()
	if err != nil {
		return []byte{}, err
	}
	if isString {
		// Return string formatted integer
		return []byte(strconv.Itoa(size)), nil
	}
	// Read string of `size`
	strBuf := make([]byte, int(size))
	if _, err := r.dbfile.Read(strBuf); err != nil {
		return []byte{}, err
	}
	return strBuf, nil
}

func (r *rdbParser) readValue(vt byte) ([]byte, error) {
	switch vt {
	case 0x0: // String Encoding
		return r.readStringEncoding()
	case 0x1: // List Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: List")
	case 0x2: // Set Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: Set")
	case 0x3: // Sorted Set Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: Sorted Set")
	case 0x4: // Hash Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: Hash")
	case 0x9: // Zipmap Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: Zipmap")
	case 0x10: // Ziplist Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: Ziplist")
	case 0x11: // Intset Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: Intset")
	case 0x12: // Sorted Set in Ziplist Encoding
		return []byte{}, fmt.Errorf("unimplemented encoding: Sorted Set in Ziplist")
	case 0x13: // Hashmap in Ziplist Encoding (Introduced in RDB version 4)
		return []byte{}, fmt.Errorf("unimplemented encoding: Hashmap in Ziplist")
	case 0x14: // List in Quicklist encoding (Introduced in RDB version 7)
		return []byte{}, fmt.Errorf("unimplemented encoding: List in Quicklist")

	default:
		return []byte{}, fmt.Errorf("unrecognized value type: %#v", vt)
	}
}
