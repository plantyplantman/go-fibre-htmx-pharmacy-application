package bigc

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
)

func GetUrl(baseUrl string, apiUrl string, params map[string]string) string {
	return fmt.Sprintf("%s%s%s", baseUrl, apiUrl, FormatUrlParams(params))
}

func FormatUrlParams(params map[string]string) string {
	if len(params) < 1 {
		return ""
	}

	values := url.Values{}
	for key, value := range params {
		values.Add(key, value)
	}

	return "?" + values.Encode()
}

func WriteToTsv(path string, headers []string, content [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		fmt.Println(e)
	}
	w := csv.NewWriter(f)
	w.Comma = '\t'
	defer w.Flush()

	e = w.Write(headers)
	if e != nil {
		fmt.Println(e)
	}

	for _, row := range content {
		w.Write(row)
	}
	return e
}

func GetStructFields(s any) ([]string, error) {
	if s == nil {
		return nil, errors.New("NO FIELD FOR A NIL DIMWIT")
	}
	v := reflect.ValueOf(s)
	var r []string
	for i := 0; i < v.NumField(); i++ {
		r = append(r, v.Type().Field(i).Name)
	}
	return r, nil
}

func GetStructValues(s any) ([]reflect.Value, error) {
	if s == nil {
		return nil, errors.New("NO FIELD FOR A NIL DIMWIT")
	}
	v := reflect.ValueOf(s)
	var r []reflect.Value
	for i := 0; i < v.NumField(); i++ {
		r = append(r, v.Field(i))
	}
	return r, nil
}

func RemoveSubstring(s string, t string) string {
	if strings.Contains(s, t) {
		ss := strings.Split(s, t)
		if len(ss) > 1 {
			return ss[1]
		}
	}
	return s
}

func CleanBigCommerceSku(sku string) (string, error) {
	pattern := `^\/{0,3}(\d+)(?:-[^-]*)?`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	sku = strings.TrimSpace(sku)
	matches := re.FindStringSubmatch(sku)
	if len(matches) < 1 {
		return "", errors.New("Invalid sku: " + sku)
	}

	s := matches[1]

	for strings.HasPrefix(s, "0") {
		s = strings.TrimPrefix(s, "0")
	}

	return s, nil
}
