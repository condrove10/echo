package echo

import (
	"github.com/go-playground/validator/v10"
	"github.com/traefik/paerser/parser"
)

type Configuration struct {
	Services map[string]Service `validate:"omitempty,dive,required"`
}

type Service struct {
	Enable   bool   `validate:"omitempty"`
	Hostname string `validate:"required"`
	Network  string `validate:"required,min=1,max=100"`
	Api      api    `validate:"required"`
}

type api struct {
	Name   string `validate:"required,min=1,max=100"`
	Client client `validate:"required"`
}

type client struct {
	Address       string `validate:"required,min=1,max=100"`
	Authorization string `validate:"omitempty,max=10000"`
	Tls           tls    `validat:"omitempty"`
}

type tls struct {
	Skip bool `validate:"omitempty"`
}

func Decode(labels map[string]string) (Configuration, error) {
	echo := Configuration{}

	if err := parser.Decode(labels, &echo, "echo", "echo.services."); err != nil {
		return Configuration{}, err
	}

	if err := validator.New().Struct(&echo); err != nil {
		return Configuration{}, err
	}

	return echo, nil
}
