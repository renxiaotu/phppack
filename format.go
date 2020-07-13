package phppack

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var formatCache = make(map[string][]packType)

func parsePackFormats(f string) ([]packType, error) {
	if cached, ok := formatCache[f]; ok {
		return cached, nil
	}
	pts := make([]packType, 0)
	//第一位只能是格式
	if !strings.Contains(formatOptions, f[0:1]) {
		return nil, errors.New("format error")
	}

	//检查是否有不支持的字符
	reg := regexp.MustCompile(regexpFormatOptions)
	if reg.MatchString(f) {
		return nil, errors.New("format contains characters that are not supported")
	}

	//*号后面只可能是不需要传参的xX@
	ind := strings.LastIndex(f, "*")
	if ind > -1 && len(f) > ind+1 {
		reg = regexp.MustCompile("[^xX@]]")
		if reg.MatchString(f[ind+1:]) {
			return nil, errors.New("'*' can only be followed by 'x/X/@'")
		}
	}

	//解析格式
	pt := packType{}
	err := errors.New("")
	for len([]byte(f)) > 0 {
		pt, err = nextPackFormatTypes(&f)
		if err != nil {
			return nil, err
		}
		pts = append(pts, pt)
	}
	formatCache[f] = pts
	return pts, nil
}

func nextPackFormatTypes(f *string) (packType, error) {
	var (
		pt  = packType{Name: "", Type: nil, tag: packTag{Type: (*f)[:1], Size: 1}}
		tag = &(pt.tag)
		reg = regexp.MustCompile("[" + formatOptions + "]")
		loc = reg.FindStringIndex((*f)[1:])
		err = errors.New("")
		num = ""
	)
	if len(loc) > 0 {
		num = (*f)[1 : loc[0]+1]
		if num == "*" {
			tag.Size = -1
		} else if num == "" {
			tag.Size = 1
		} else {
			tag.Size, err = strconv.Atoi(num)
			if err != nil {
				return pt, err
			}
		}
		*f = (*f)[loc[0]+1:]
	} else {
		if len(*f) > 1 {
			num = (*f)[1:]
			if num == "*" {
				tag.Size = -1
			} else {
				tag.Size, err = strconv.Atoi(num)
				if err != nil {
					return pt, err
				}
			}
		} else {
			tag.Size = 1
		}
		*f = ""
	}
	if tag.Size == 0 {
		return pt, errors.New("the number of parameters cannot be 0")
	}
	if tag.Size == -1 && tag.Type == "@" {
		return pt, errors.New("'@' cannot be followed by '*'")
	}

	//除了aAhH，其它类型的*号只能在末尾
	if tag.Size == -1 && *f != "" && !strings.Contains(stringFormatOptions, tag.Type) {
		return pt, errors.New("except for aAhH, other types of '*' can only be at the end")
	}
	return pt, nil
}

var formatCacheUn = make(map[string][]packType)

func parseUnPackFormats(f string) ([]packType, error) {
	if cached, ok := formatCacheUn[f]; ok {
		return cached, nil
	}
	pts := make([]packType, 0)

	//解析格式
	pt := packType{}
	err := errors.New("")
	fs := strings.Split(f, "/")
	for len(fs) > 0 {
		pt, err = nextUnPackFormatTypes(fs[0])
		if err != nil {
			return nil, err
		}
		pts = append(pts, pt)
		fs = fs[1:]
	}
	formatCacheUn[f] = pts
	return pts, nil
}

func nextUnPackFormatTypes(f string) (packType, error) {
	var (
		pt  = packType{Name: "", Type: nil, tag: packTag{Type: f[:1], Size: 1}}
		tag = &(pt.tag)
		reg = regexp.MustCompile("[^0-9*]")
		loc = reg.FindStringIndex((f)[1:])
		err = errors.New("")
	)
	//第一位只能是格式
	if !strings.Contains(formatOptions, pt.tag.Type) {
		return pt, errors.New(f + "format error")
	}

	if len(loc) == 0 {
		if len([]byte(f)) > 1 {
			if f[1:] == "*" {
				tag.Size = -1
			} else {
				tag.Size, err = strconv.Atoi(f[1:])
				if err != nil {
					return pt, err
				}
			}

		} else {
			tag.Size = 1
		}
	} else {
		if loc[0] == 0 {
			tag.Size = 1
			pt.Name = f[1:]
		} else {
			if f[1:] == "*" {
				tag.Size = -1
			} else {
				tag.Size, err = strconv.Atoi(f[1 : loc[0]+1])
				if err != nil {
					return pt, err
				}
			}
			pt.Name = f[loc[0]+1:]
		}
	}

	if tag.Size == 0 {
		return pt, errors.New("the number of parameters cannot be 0")
	}

	return pt, nil
}
