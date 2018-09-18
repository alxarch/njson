package njsontest

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

// Option is a test case option.
type Option interface {
	set(t *T)
}

type option func(t *T)

func (f option) set(t *T) {
	f(t)
}

// Method sets a custom method name for the test case.
func Method(methodName string) Option {
	return option(func(t *T) {
		t.method = methodName
	})
}

// Error sets an error to expect
func Error(err interface{}) Option {
	var check func(error) error
	switch e := err.(type) {
	case nil:
		check = nil
	case error:
		if e != nil {
			check = func(err error) error {
				if err == e {
					return nil
				}
				return err
			}
		}
	case func(error) error:
		check = e
	case string:
		if e != "" {
			check = func(err error) error {
				if strings.Contains(err.Error(), e) {
					return nil
				}
				return err
			}

		}
	case *regexp.Regexp:
		if e != nil {
			check = func(err error) error {
				if e.MatchString(err.Error()) {
					return nil
				}
				return err
			}

		}
	case bool:
		if e {
			check = func(err error) error {
				if err == nil {
					return errors.New("Expecting error")
				}
				return nil
			}
		}
	default:
		panic("Invalid Error option")
	}

	return option(func(t *T) {
		t.check = check
	})
}

func defaultJSON(x interface{}) Option {
	return option(func(t *T) {
		t.json = func() ([]byte, error) {
			return json.Marshal(x)
		}
	})
}

func Value(x interface{}) Option {
	return option(func(t *T) {
		t.value = x
	})
}

// JSON sets JSON data for the test case.
func JSON(x interface{}) Option {
	var getjson func() ([]byte, error)
	switch j := x.(type) {
	case func() ([]byte, error):
		getjson = j
	case []byte:
		getjson = func() ([]byte, error) {
			return j, nil
		}
	case string:
		getjson = func() ([]byte, error) {
			return []byte(j), nil
		}
	case io.Reader:
		getjson = func() ([]byte, error) {
			return ioutil.ReadAll(j)
		}
	case nil:
		getjson = func() ([]byte, error) {
			return []byte(`null`), nil
		}
	default:
		getjson = func() ([]byte, error) {
			return json.Marshal(x)
		}
	}
	return option(func(t *T) {
		t.json = getjson
	})
}
