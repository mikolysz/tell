package main

import (
	"strings"
	"testing"
)

func TestValidation(t *testing.T) {
	valid := []string{
		"hello",
		"-t token",
		"-a",
	}

	invalid := []string{
		"",
		"-t token hello",
		"-a hello",
	}

	for _, v := range valid {
		t.Run("Args: "+v, func(t *testing.T) {
			args := strings.Split(v, " ")
			env, err := initEnvironment(args)
			if err != nil {
				t.Errorf("error while initializing environment: %v", err)
			}

			if err := env.validate(); err != nil {
				t.Errorf("validation failed for valid arguments: %s: %v", v, err)
			}
		})
	}

	for _, iv := range invalid {
		t.Run("Args: "+iv, func(t *testing.T) {
			args := strings.Split(iv, " ")
			env, err := initEnvironment(args)
			if err != nil {
				t.Errorf("error while initializing environment: %v", err)
			}

			if err := env.validate(); err == nil {
				t.Errorf("validation succeeded for invalid arguments: %s", iv)
			}
		})
	}
}
