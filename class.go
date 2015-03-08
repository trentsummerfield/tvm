package main

import (
	"bytes"
	"encoding/binary"
)

type Class struct {
	magic                      uint32
	minorVersion, majorVersion uint16
	constantPoolCount          uint16
}

type ParseError struct {
	reason string
}

func (e ParseError) Error() string {
	return e.reason
}

func parse(b []byte) (Class, error) {
	var magic uint32
	var minorVersion, majorVersion uint16
	var constantPoolCount uint16
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &magic)
	if err != nil {
		return Class{}, err
	}
	err = binary.Read(buf, binary.BigEndian, &minorVersion)
	if err != nil {
		return Class{}, err
	}
	err = binary.Read(buf, binary.BigEndian, &majorVersion)
	if err != nil {
		return Class{}, err
	}
	err = binary.Read(buf, binary.BigEndian, &constantPoolCount)
	if err != nil {
		return Class{}, err
	}
	return Class{
		magic:             magic,
		minorVersion:      minorVersion,
		majorVersion:      majorVersion,
		constantPoolCount: constantPoolCount,
	}, nil
}
