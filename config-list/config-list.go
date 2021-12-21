/*
 * Create time: 2021-12-21
 * Update time: 2021-12-21
 */
package config_list

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"sort"
)

type Conf struct {
	filename string
	data     []string
	reg      *regexp.Regexp
}

func Parse(expr, filename string) ([]string, error) {

	d := &Conf{filename: filename}
	if err := d.everyLineString(); err != nil {
		return nil, err
	}

	var err error
	d.reg, err = regexp.CompilePOSIX(expr)
	if err != nil {
		return nil, err
	}
	d.everyLineRegex()

	sort.Strings(d.data)
	return unique(d.data), nil
}

func (c *Conf) everyLineString() error {

	fd, err := os.Open(c.filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	br := bufio.NewReader(fd)
	for {
		a, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		c.data = append(c.data, string(a))
	}
	return nil
}

func (c *Conf) everyLineRegex() {

	var _data []string
	for i := range c.data {
		ok := c.reg.MatchString(c.data[i])
		if ok {
			_data = append(_data, c.data[i])
		}
	}
	c.data = _data
}

func unique(slice []string) []string {

	uniqMap := make(map[string]struct{})
	for _, v := range slice {
		uniqMap[v] = struct{}{}
	}

	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	return uniqSlice
}
