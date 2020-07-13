package phppack

const Version = "1.1.0"
const PackageName = "phppack"
const TagName = "pack"
const formatOptions = "aAcCdeEfgGhHiIJlLnNPqQsSvVxXZ@"
const stringFormatOptions = "aAhH"
const regexpFormatOptions = "[^" + formatOptions + "0-9*]+"

//各系统类型默认pack类型
func autoType(t string) string {
	m := make(map[string]string, 0)
	m["int8"] = "c"
	m["byte"] = "c"
	m["uint8"] = "C"
	m["int16"] = "s"
	m["uint16"] = "S"
	m["int"] = "i"
	m["uint"] = "I"
	m["int32"] = "l"
	m["uint32"] = "L"
	m["int64"] = "q"
	m["uint64"] = "Q"
	m["float32"] = "f"
	m["float64"] = "d"
	v, ok := m[t]
	if ok {
		return v
	}
	return ""
}
