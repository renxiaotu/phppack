package phppack

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var structCache = make(map[reflect.Type][]packType)
var structCacheLock sync.RWMutex
var parseLock sync.Mutex

//tag结构
func parsePackTag(tag reflect.StructTag) packTag {
	t := packTag{Type: "", Size: 1}
	s := tag.Get(TagName)
	if s == "" {
		return t
	}
	t.Type = s[:1]
	a := s[1:]
	switch a {
	case "":
		t.Size = 1
		break
	case "*":
		t.Size = -1
		break
	default:
		i, err := strconv.Atoi(a)
		if err != nil {
			t.Size = 1
		}
		t.Size = i
	}
	return t
}

//解析结构
func parseTypesLocked(v reflect.Value) ([]packType, error) {
	//需要重复这个逻辑，因为下面的parseFields（）由于锁定而不能被递归调用
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	if v.NumField() < 1 {
		return nil, errors.New(PackageName + ": Struct has no fields")
	}
	pts := make([]packType, 0)
	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}
		field := t.Field(i)
		v.Field(i)
		tag := parsePackTag(field.Tag)
		if tag.Type == "" {
			tag.Type = autoType(field.Type.Name())
			if tag.Type == "" {
				return nil, errors.New(PackageName + ":'" + field.Name + "' does not specify the format")
			}
		}
		pt := packType{
			Name: field.Name,
			Type: field.Type,
			tag:  tag,
		}
		pts = append(pts, pt)

	}
	return pts, nil
}

//结构缓存获取
func typeCacheLookup(t reflect.Type) []packType {
	structCacheLock.RLock()
	defer structCacheLock.RUnlock()
	if cached, ok := structCache[t]; ok {
		return cached
	}
	return nil
}

//解析结构
func parseTypes(v reflect.Value) ([]packType, error) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	//读取解析缓存
	if cached := typeCacheLookup(t); cached != nil {
		return cached, nil
	}

	//全局锁
	parseLock.Lock()
	defer parseLock.Unlock()

	//再次检查缓存，以防parseLock刚刚被释放
	if cached := typeCacheLookup(t); cached != nil {
		return cached, nil
	}

	//开始分析缓存
	pts, err := parseTypesLocked(v)
	if err != nil {
		return nil, err
	}

	//struct模式的数据number型不允许有*
	for i := 0; i < len(pts); i++ {
		if pts[i].tag.Size == -1 && strings.Contains("[^aAhH]", pts[i].tag.Type) {
			return nil, errors.New(PackageName + ":" + pts[i].tag.Type + " does not accept * sign")
		}
	}

	structCacheLock.Lock()
	structCache[t] = pts
	structCacheLock.Unlock()
	return pts, nil
}
