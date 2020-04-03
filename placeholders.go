/*
Copyright 2019 - The TXTDirect Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package txtdirect

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// PlaceholderRegex finds the placeholders like {x}
var PlaceholderRegex = regexp.MustCompile("{[~>?]?\\w+}")

// parsePlaceholders gets a string input and looks for placeholders inside
// the string. it will then replace them with the actual data from the request
func parsePlaceholders(input string, r *http.Request, pathSlice []string) (string, error) {
	placeholders := PlaceholderRegex.FindAllStringSubmatch(input, -1)
	for _, placeholder := range placeholders {
		switch placeholder[0] {
		case "{uri}":
			input = strings.Replace(input, "{uri}", r.URL.RequestURI(), -1)
		case "{dir}":
			dir, _ := path.Split(r.URL.Path)
			input = strings.Replace(input, "{dir}", dir, -1)
		case "{file}":
			_, file := path.Split(r.URL.Path)
			input = strings.Replace(input, "{file}", file, -1)
		case "{host}":
			input = strings.Replace(input, "{host}", r.Host, -1)
		case "{hostonly}":
			// Removes port from host
			host := r.Host
			if strings.Contains(r.Host, ":") {
				hostSlice := strings.Split(r.Host, ":")
				host = hostSlice[0]
			}
			input = strings.Replace(input, "{hostonly}", host, -1)
		case "{method}":
			input = strings.Replace(input, "{method}", r.Method, -1)
		case "{path}":
			input = strings.Replace(input, "{path}", r.URL.Path, -1)
		case "{path_escaped}":
			input = strings.Replace(input, "{path_escaped}", url.QueryEscape(r.URL.Path), -1)
		case "{port}":
			input = strings.Replace(input, "{port}", r.URL.Port(), -1)
		case "{query}":
			input = strings.Replace(input, "{query}", r.URL.RawQuery, -1)
		case "{query_escaped}":
			input = strings.Replace(input, "{query_escaped}", url.QueryEscape(r.URL.RawQuery), -1)
		case "{uri_escaped}":
			input = strings.Replace(input, "{uri_escaped}", url.QueryEscape(r.URL.RequestURI()), -1)
		case "{user}":
			user, _, ok := r.BasicAuth()
			if !ok {
				input = strings.Replace(input, "{user}", "", -1)
			}
			input = strings.Replace(input, "{user}", user, -1)
		}
		// For multi-level tlds such as "example.co.uk", "co" would be used as {label2},
		// "example" would be {label1} and "uk" would be {label3}
		if strings.HasPrefix(placeholder[0], "{label") {
			nStr := placeholder[0][6 : len(placeholder[0])-1] // get the integer N in "{labelN}"
			n, err := strconv.Atoi(nStr)
			if err != nil {
				return "", err
			}
			if n < 1 {
				return "", fmt.Errorf("{label0} is not supported")
			}
			// Removes port from host
			host := r.Host
			if strings.Contains(r.Host, ":") {
				hostSlice := strings.Split(r.Host, ":")
				host = hostSlice[0]
			}
			labels := strings.Split(host, ".")
			if n > len(labels) {
				return "", fmt.Errorf("Cannot parse a label greater than %d", len(labels))
			}
			input = strings.Replace(input, placeholder[0], labels[n-1], -1)
		}
		if placeholder[0][1] == '>' {
			want := placeholder[0][2 : len(placeholder[0])-1]
			for key, values := range r.Header {
				// Header placeholders (case-insensitive)
				if strings.EqualFold(key, want) {
					input = strings.Replace(input, placeholder[0], strings.Join(values, ","), -1)
				}
			}
		}
		if placeholder[0][1] == '~' {
			name := placeholder[0][2 : len(placeholder[0])-1]
			if cookie, err := r.Cookie(name); err == nil {
				input = strings.Replace(input, placeholder[0], cookie.Value, -1)
			}
		}
		if placeholder[0][1] == '?' {
			query := r.URL.Query()
			name := placeholder[0][2 : len(placeholder[0])-1]
			input = strings.Replace(input, placeholder[0], query.Get(name), -1)
		}

		// Numbered Regex matches
		if regexp.MustCompile("(\\d+)").MatchString(string(placeholder[0][1])) {
			matches := r.Context().Value("regexMatches")
			index, err := strconv.Atoi(string(placeholder[0][1]))
			if err != nil {
				return "", fmt.Errorf("couldn't get index of regex match")
			}
			input = strings.Replace(input, placeholder[0],
				reflect.ValueOf(matches).Index(index-1).String(), -1)
		}

		// Named regex matches
		if regexp.MustCompile("([a-zA-Z]+[0-9]*)").MatchString(string(placeholder[0][1 : len(placeholder[0])-1])) {
			matches := r.Context().Value("regexMatches")
			mapReflect := reflect.ValueOf(matches)
			if mapReflect.Kind() == reflect.Map {
				iterator := reflect.ValueOf(matches).MapRange()
				for iterator.Next() {
					if iterator.Key().String() == string(placeholder[0][1:len(placeholder[0])-1]) {
						input = strings.Replace(input, placeholder[0], iterator.Value().String(), -1)
					}
				}
			}
		}
	}

	for k, v := range pathSlice {
		input = strings.Replace(input, fmt.Sprintf("{$%d}", k+1), v, -1)
	}

	return input, nil
}
