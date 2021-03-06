package app

import (
	"errors"
	"fmt"
)

type Status int

const (
	Correct = iota
	Malicious
)

func (t Status) String() string {
	switch t {
	case Correct:
		return "Correct"
	case Malicious:
		return "Malicious"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

func ParseStatus(str string) (status Status, err error) {
	switch str {
	case "Correct":
		return Correct, nil
	case "Malicious":
		return Malicious, nil
	default:
		return -1, errors.New("cannot parse Status")
	}
}
