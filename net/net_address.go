package net

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	StringToAddressConversionError = "could not convert invalid string to address structure"
)

func NewAddress(ip string) (*Address, error) {
	if !strings.Contains(ip, ":") {
		return nil, errors.New(StringToAddressConversionError)
	}
	split := strings.Split(ip, ":")
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, errors.New(StringToAddressConversionError)
	}

	address := new(Address)
	address.Host = split[0]
	address.Port = port

	return address, nil
}

func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
