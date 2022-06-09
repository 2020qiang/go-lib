/*
 * config/ini
 *
 * create time 2022-01-21
 * update time 2022-03-16
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

type errsT struct {
	data []error
}

func (errs *errsT) add(err error) {
	errs.data = append(errs.data, err)
}
func (errs *errsT) addFormat(format string, a ...interface{}) {
	errs.add(fmt.Errorf(format, a...))
}

type File struct {
	cfg      *iniV1.File
	errs     errsT
	_default *fileDefault
}

func Load(name string, _default bool) *File {

	if len(name) == 0 {
		name = defaultFilename()
	}

	var data File
	cfg, err := iniV1.Load(name)
	if err != nil {
		data.errs.add(err)
	}
	data.cfg = cfg
	data._default = &fileDefault{enable: _default, cfg: cfg}
	errs := data._default.init(len(data.errs.data) != 0)
	for i := range errs {
		data.errs.add(errs[i])
	}
	return &data
}

func (f *File) Errors() []error {
	return f.errs.data
}

func (f *File) Sections() []string {
	if len(f.errs.data) != 0 {
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
	if len(f.errs.data) != 0 {
		return nil
	}

	switch sectionString {
	case defaultSection:
	case defaultSectionRaw:
		f.errs.addFormat("%s.*  keys not allow", sectionString)
		return nil
	}

	keys := f.cfg.Section(sectionString).KeyStrings()
	for key := range f._default.values {
		keys = append(keys, key)
	}
	return uniq(keys)
}

// 返回值 valueString success
func (f *File) value(sectionString, key string) string {
	if f.cfg == nil {
		return ""
	}
	str := f.cfg.Section(sectionString).Key(key).Value()

	// 存在
	if len(str) != 0 {
		return str
	}

	// 不存在但有有效的默认值
	if len(str) == 0 && f._default.active(sectionString, key) {
		return f._default.value(sectionString, key)
	}

	// 不存在且没有默认值
	if len(str) == 0 && !f._default.active(sectionString, key) {
		f.errs.addFormat("%s.%s does not exist and has no default value", sectionString, key)
		return ""
	}

	// 未知错误
	f.errs.addFormat("%s.%s unknown error", sectionString, key)
	return ""
}

func (f *File) String(sectionString, key string) string {
	return f.value(sectionString, key)
}

func (f *File) Int(sectionString, key string) int {
	str := f.value(sectionString, key)
	if len(str) == 0 {
		return 0
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		f.errs.addFormat("%s.%s %s", sectionString, key, err)
		return 0
	}
	return v
}

func (f *File) Float64(sectionString, key string) float64 {
	str := f.value(sectionString, key)
	if len(str) == 0 {
		return 0
	}
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		f.errs.addFormat("%s.%s %s", sectionString, key, err)
		return 0
	}
	return v
}

func (f *File) Duration(sectionString, key string) time.Duration {
	str := f.value(sectionString, key)
	if len(str) == 0 {
		return 0
	}
	v, err := time.ParseDuration(str)
	if err != nil {
		f.errs.addFormat("%s.%s %s", sectionString, key, err)
		return 0
	}
	return v
}

func (f *File) Bool(sectionString, key string) bool {
	str := f.value(sectionString, key)
	if len(str) == 0 {
		f.errs.addFormat("%s.%s %s", sectionString, key, "value is empty")
		return false
	}

	if str == "true" {
		return true
	}
	if str == "false" {
		return false
	}

	f.errs.addFormat("%s.%s %s", sectionString, key, "non-valid value")
	return false
}

type fileDefault struct {
	enable bool
	cfg    *iniV1.File
	values map[string]string
}

// 初始化获取默认值
func (sd *fileDefault) init(jumpover bool) []error {
	if jumpover {
		return nil
	}
	var errs []error
	cfg := sd.cfg.Section(defaultSection)
	if sd.enable && cfg != nil {
		sd.values = make(map[string]string)
		keys := cfg.Keys()
		for i := range keys {
			v := keys[i].Value()
			if len(v) == 0 {
				errs = append(errs, fmt.Errorf("default.%s string is empty", keys[i].Name()))
				continue
			}
			sd.values[keys[i].Name()] = v
		}
	}
	return errs
}

// 默认值 [default] 是否已配置上并可用
func (sd *fileDefault) active(sectionString, key string) bool {
	section := sd.enable && sd.cfg != nil && sectionString != globalSection && sectionString != defaultSection
	_, ok := sd.values[key]
	return section && ok
}

// 默认值的有效值
func (sd *fileDefault) value(sectionString, key string) string {
	return sd.values[key]
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
