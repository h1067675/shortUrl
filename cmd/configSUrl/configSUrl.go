package configSUrl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var Config struct {
	NetAddressServerShortener
	NetAddressServerExpand
}

type NetAddressServerShortener struct {
	Host string
	Port int
}

type NetAddressServerExpand struct {
	Host string
	Port int
}

// checkNetAddress - функция проверяющая на корректность указания пары host:port и в случае ошибки передающей значения по умолчанию
func checkNetAddress(s string) (host string, port int, e error) {
	v := strings.Split(s, "://")
	if len(v) != 2 {
		e = errors.New("Incorrect net address.")
		return "localhost", 8080, e
	}
	a := strings.Split(v[1], ":")
	if len(a) < 1 || len(a) > 2 {
		e = errors.New("Incorrect net address.")
		return "localhost", 8080, e
	}
	host = a[0]
	if a[1] != "" {
		port, e = strconv.Atoi(a[1])
		if e != nil {
			return "localhost", 8080, e
		}
	} else {
		port = 80
	}
	return
}

func (n *NetAddressServerShortener) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

func (n *NetAddressServerShortener) Set(flagValue string) (err error) {
	n.Host, n.Port, err = checkNetAddress(flagValue)
	if err != nil {
		return err
	}
	return nil
}

func (n *NetAddressServerExpand) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

func (n *NetAddressServerExpand) Set(flagValue string) (err error) {
	n.Host, n.Port, err = checkNetAddress(flagValue)
	if err != nil {
		return err
	}
	return nil
}
