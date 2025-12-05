// Flexible ID type for JSON number or string forms.

package helpers

import (
    "encoding/json"
    "strconv"
    "strings"
)

// FlexID 支持 JSON 中数字或字符串两种形式的无符号 ID
type FlexID uint64

func (f *FlexID) UnmarshalJSON(b []byte) error {
    s := strings.TrimSpace(string(b))
    if s == "" || s == "null" {
        *f = 0
        return nil
    }
    // 尝试解析为字符串形式
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
        unq, err := strconv.Unquote(s)
        if err != nil { return err }
        v, err := strconv.ParseUint(unq, 10, 64)
        if err != nil { return err }
        *f = FlexID(v)
        return nil
    }
    // 尝试解析为数字形式
    var v uint64
    if err := json.Unmarshal(b, &v); err != nil { return err }
    *f = FlexID(v)
    return nil
}

// Uint64 返回原始 uint64 值
func (f FlexID) Uint64() uint64 { return uint64(f) }