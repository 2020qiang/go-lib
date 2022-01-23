/*
 * config/ini
 *
 * create time 2022-01-21
 * update time 2022-01-23
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
	data._default.init(len(data.errs) != 0)
	return &data
}

func (f *File) SectionStrings() []string {
	if len(f.errs) != 0 {
		return nil
	}
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
	if len(f.errs) != 0 {
		return nil
	}
	rawKeys := f.cfg.Section(sectionString).KeyStrings()
	if !f._default.enable(sectionString) {
		return rawKeys
	}
	for k := range f._default.rawData {
		rawKeys = append(rawKeys, k)
	}
	return uniq(rawKeys)
}

func (f *File) value(sectionString, key string) string {
	if len(f.errs) != 0 {
		return ""
	}
	v := f.cfg.Section(sectionString).Key(key).Value()
	if len(v) == 0 {
		if f._default.enable(sectionString) {
			return f._default.defaultString(sectionString, key)
		}
		f.errs = append(f.errs, fmt.Errorf("%s.%s the original value is empty", sectionString, key))
		return ""
	}
	return v
}

func (f *File) String(sectionString, key string) string {
	return f.value(sectionString, key)
}

func (f *File) Int(sectionString, key string) int {
	str := f.value(sectionString, key)
	v, err := strconv.Atoi(str)
	if err != nil {
		f.errs = append(f.errs, fmt.Errorf("%s.%s %s", sectionString, key, err))
		return 0
	}
	return v
}

func (f *File) Duration(sectionString, key string) time.Duration {
	str := f.value(sectionString, key)
	v, err := time.ParseDuration(str)
	if err != nil {
		f.errs = append(f.errs, fmt.Errorf("%s.%s %s", sectionString, key, err))
		return 0
	}
	return v
}

func (f *File) Errors() []error {
	return f.errs
}

type sectionDefault struct {
	rawEnable bool
	cfg       *iniV1.File
	rawData   map[string]string
}

func (sd *sectionDefault) init(jumpover bool) {
	if jumpover {
		return
	}
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

func (sd *sectionDefault) defaultString(sectionString, key string) string {
	if sectionString != defaultSection && sectionString != globalSection {
		return sd.rawData[key]
	}
	return ""
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
