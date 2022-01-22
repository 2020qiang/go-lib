/*
 * config/ini
 *
 * create time 2022-01-22
 * update time 2022-01-22
 */

package ini

import (
	"fmt"
	iniV1 "gopkg.in/ini.v1"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	globalSection     = "global"
	defaultSection    = "default"
	defaultSectionRaw = "DEFAULT"
)

type File struct {
	cfg      *iniV1.File
	errs     []error
	_default *sectionDefault
}

func Load(name string, _default bool) *File {

	if len(name) == 0 {
		name = defaultFilename()
	}

	var data File
	cfg, err := iniV1.Load(name)
	if err != nil {
		data.errs = append(data.errs, err)
	}
	data.cfg = cfg
	data._default = &sectionDefault{rawEnable: _default, cfg: cfg}
	data._default.init()
	return &data
}

func (f *File) SectionStrings() []string {
	v := f.cfg.SectionStrings()
	var data []string
	for i := range v {
		switch v[i] {
		case globalSection:
		case defaultSection:
		case defaultSectionRaw:
		default:
			data = append(data, v[i])
		}
	}
	return data
}

func (f *File) Keys(sectionString string) []string {
	rawKeys := f.cfg.Section(sectionString).KeyStrings()
	if !f._default.enable(sectionString) {
		return rawKeys
	}
	for k := range f._default.rawData {
		rawKeys = append(rawKeys, k)
	}
	return uniq(rawKeys)
}

func (f *File) String(sectionString, key string) string {
	data := f.cfg.Section(sectionString).Key(key).String()
	if len(data) == 0 && !f._default.enable(sectionString) {
		f.errs = append(f.errs, fmt.Errorf("%s.%s string does not exist or is empty", sectionString, key))
		return ""
	}
	if len(data) == 0 && f._default.enable(sectionString) {
		return f._default.tryString(sectionString, key)
	}
	return data
}

func (f *File) Int(sectionString, key string) int {
	data, err := f.cfg.Section(sectionString).Key(key).Int()
	if err != nil && !f._default.enable(sectionString) {
		f.errs = append(f.errs, fmt.Errorf("%s.%s string does not exist or is empty", sectionString, key))
		return 0
	}
	if err != nil {
		data, err = f._default.tryInt(sectionString, key)
		if err != nil {
			f.errs = append(f.errs, fmt.Errorf("%s.%s %s", sectionString, key, err))
			return 0
		}
	}
	return data
}

func (f *File) Duration(sectionString, key string) time.Duration {
	data, err := f.cfg.Section(sectionString).Key(key).Duration()
	if err != nil && !f._default.enable(sectionString) {
		f.errs = append(f.errs, fmt.Errorf("%s.%s %s", sectionString, key, err))
		return 0
	}
	if err != nil {
		data, err = f._default.tryDuration(sectionString, key)
		if err != nil {
			f.errs = append(f.errs, fmt.Errorf("%s.%s %s", sectionString, key, err))
			return 0
		}
	}
	return data
}

func (f *File) Errors() []error {
	return f.errs
}

type sectionDefault struct {
	rawEnable bool
	cfg       *iniV1.File
	rawData   map[string]string
}

func (sd *sectionDefault) init() {
	cfg := sd.cfg.Section(defaultSection)
	if sd.rawEnable && cfg != nil {
		sd.rawData = make(map[string]string)
		keys := cfg.Keys()
		for i := range keys {
			sd.rawData[keys[i].Name()] = keys[i].Value()
		}
	}
}

func (sd *sectionDefault) enable(sectionString string) bool {
	return sd.rawEnable && sd.cfg != nil && sectionString != globalSection
}

func (sd *sectionDefault) tryString(sectionString, key string) string {
	if sectionString != defaultSection && sectionString != globalSection {
		return sd.rawData[key]
	}
	return ""
}

func (sd *sectionDefault) tryInt(sectionString, key string) (int, error) {
	if sectionString != defaultSection && sectionString != globalSection {
		v, err := strconv.ParseInt(sd.rawData[key], 0, 64)
		return int(v), err
	}
	return 0, nil
}

func (sd *sectionDefault) tryDuration(sectionString, key string) (time.Duration, error) {
	if sectionString != defaultSection && sectionString != globalSection {
		return time.ParseDuration(sd.rawData[key])
	}
	return 0, nil
}

// 配置文件名字
func defaultFilename() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s.ini", strings.Split(os.Args[0], ".exe")[0])
	}
	return fmt.Sprintf("%s.ini", os.Args[0])
}

// 数组去重
func uniq(list []string) []string {
	// 创建一个临时map用来存储数组元素
	temp := make(map[string]bool)
	for i, v := range list {
		// 遍历数组元素，判断此元素是否已经存在map中
		_, ok := temp[v]
		if ok {
			list = append(list[:i], list[i+1:]...)
		} else {
			temp[v] = true
		}
	}
	return list
}
