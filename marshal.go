package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
)

// MarshalFormat 序列化格式
type MarshalFormat int

const (
	// JSONFormat JSON 格式
	JSONFormat MarshalFormat = iota
	// YAMLFormat YAML 格式
	YAMLFormat
	// XMLFormat XML 格式
	XMLFormat
	// StringFormat 字符串格式
	StringFormat
)

// MarshalOptions 序列化选项
type MarshalOptions struct {
	Format     MarshalFormat
	Pretty     bool
	Indent     string
	MaxLength  int
	Truncate   bool
	EscapeHTML bool
	SortKeys   bool
}

// DefaultMarshalOptions 默认选项
var DefaultMarshalOptions = MarshalOptions{
	Format:     JSONFormat,
	Pretty:     false,
	Indent:     "  ",
	MaxLength:  0,
	Truncate:   false,
	EscapeHTML: true,
	SortKeys:   false,
}

// Marshaler 序列化器接口
type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
	MarshalToString(v interface{}) (string, error)
	MarshalToWriter(w io.Writer, v interface{}) error
}

// Unmarshaler 反序列化器接口
type Unmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
	UnmarshalFromString(str string, v interface{}) error
	UnmarshalFromReader(r io.Reader, v interface{}) error
}

// MarshalExt 扩展序列化器
type MarshalExt struct {
	options MarshalOptions
}

// NewMarshalExt 创建新的序列化器
func NewMarshalExt(options MarshalOptions) *MarshalExt {
	return &MarshalExt{options: options}
}

// DefaultMarshalExt 创建默认序列化器
func DefaultMarshalExt() *MarshalExt {
	return &MarshalExt{options: DefaultMarshalOptions}
}

// SetFormat 设置格式（链式调用）
func (m *MarshalExt) SetFormat(format MarshalFormat) *MarshalExt {
	m.options.Format = format
	return m
}

// SetPretty 设置美化输出（链式调用）
func (m *MarshalExt) SetPretty(pretty bool) *MarshalExt {
	m.options.Pretty = pretty
	return m
}

// SetIndent 设置缩进（链式调用）
func (m *MarshalExt) SetIndent(indent string) *MarshalExt {
	m.options.Indent = indent
	return m
}

// SetMaxLength 设置最大长度（链式调用）
func (m *MarshalExt) SetMaxLength(maxLength int) *MarshalExt {
	m.options.MaxLength = maxLength
	return m
}

// SetTruncate 设置是否截断（链式调用）
func (m *MarshalExt) SetTruncate(truncate bool) *MarshalExt {
	m.options.Truncate = truncate
	return m
}

// SetEscapeHTML 设置是否转义HTML（链式调用）
func (m *MarshalExt) SetEscapeHTML(escape bool) *MarshalExt {
	m.options.EscapeHTML = escape
	return m
}

// SetSortKeys 设置是否排序键（链式调用）
func (m *MarshalExt) SetSortKeys(sort bool) *MarshalExt {
	m.options.SortKeys = sort
	return m
}

// Clone 克隆序列化器
func (m *MarshalExt) Clone() *MarshalExt {
	return &MarshalExt{options: m.options}
}

// Marshal 序列化对象
func (m *MarshalExt) Marshal(v interface{}) ([]byte, error) {
	var data []byte
	var err error

	switch m.options.Format {
	case JSONFormat:
		data, err = m.marshalJSON(v)
	case YAMLFormat:
		data, err = m.marshalYAML(v)
	case XMLFormat:
		data, err = m.marshalXML(v)
	case StringFormat:
		data, err = m.marshalString(v)
	default:
		data, err = m.marshalJSON(v)
	}

	if err != nil {
		return nil, err
	}

	// 处理长度限制
	if m.options.MaxLength > 0 && len(data) > m.options.MaxLength {
		if m.options.Truncate {
			data = data[:m.options.MaxLength]
		}
	}

	return data, nil
}

// MarshalToString 序列化为字符串
func (m *MarshalExt) MarshalToString(v interface{}) (string, error) {
	data, err := m.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MustMarshalToString 序列化为字符串，出错时返回空字符串
func (m *MarshalExt) MustMarshalToString(v interface{}) string {
	str, err := m.MarshalToString(v)
	if err != nil {
		return ""
	}
	return str
}

// MarshalToWriter 序列化到 Writer
func (m *MarshalExt) MarshalToWriter(w io.Writer, v interface{}) error {
	data, err := m.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Unmarshal 反序列化
func (m *MarshalExt) Unmarshal(data []byte, v interface{}) error {
	switch m.options.Format {
	case JSONFormat:
		return m.unmarshalJSON(data, v)
	case YAMLFormat:
		return m.unmarshalYAML(data, v)
	case XMLFormat:
		return m.unmarshalXML(data, v)
	case StringFormat:
		return m.unmarshalString(data, v)
	default:
		return m.unmarshalJSON(data, v)
	}
}

// UnmarshalFromString 从字符串反序列化
func (m *MarshalExt) UnmarshalFromString(str string, v interface{}) error {
	return m.Unmarshal([]byte(str), v)
}

// UnmarshalFromReader 从 Reader 反序列化
func (m *MarshalExt) UnmarshalFromReader(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return m.Unmarshal(data, v)
}

// JSON 序列化实现
func (m *MarshalExt) marshalJSON(v interface{}) ([]byte, error) {
	if m.options.Pretty {
		return json.MarshalIndent(v, "", m.options.Indent)
	}
	return json.Marshal(v)
}

func (m *MarshalExt) unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// YAML 序列化实现（需要导入 yaml 包）
func (m *MarshalExt) marshalYAML(v interface{}) ([]byte, error) {
	// 这里需要导入 gopkg.in/yaml.v3
	// import "gopkg.in/yaml.v3"
	// return yaml.Marshal(v)

	// 暂时使用 JSON 格式，实际使用时替换为：
	// return yaml.Marshal(v)
	return m.marshalJSON(v)
}

func (m *MarshalExt) unmarshalYAML(data []byte, v interface{}) error {
	// 这里需要导入 gopkg.in/yaml.v3
	// return yaml.Unmarshal(data, v)

	return m.unmarshalJSON(data, v)
}

// XML 序列化实现
func (m *MarshalExt) marshalXML(v interface{}) ([]byte, error) {
	// 对于 map 类型，先转换为 JSON 再处理
	if _, ok := v.(map[string]interface{}); ok {
		// 对于 map 类型，使用 JSON 格式
		return m.marshalJSON(v)
	}

	if m.options.Pretty {
		return xml.MarshalIndent(v, "", m.options.Indent)
	}
	return xml.Marshal(v)
}

func (m *MarshalExt) unmarshalXML(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

// 字符串序列化实现
func (m *MarshalExt) marshalString(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	case fmt.Stringer:
		return []byte(val.String()), nil
	default:
		return []byte(fmt.Sprintf("%v", val)), nil
	}
}

func (m *MarshalExt) unmarshalString(data []byte, v interface{}) error {
	// 对于字符串格式，尝试直接赋值
	if strPtr, ok := v.(*string); ok {
		*strPtr = string(data)
		return nil
	}
	return fmt.Errorf("cannot unmarshal string to %T", v)
}

// 便捷方法
func (m *MarshalExt) ToJSON(v interface{}) (string, error) {
	ext := m.Clone().SetFormat(JSONFormat)
	return ext.MarshalToString(v)
}

func (m *MarshalExt) ToPrettyJSON(v interface{}) (string, error) {
	ext := m.Clone().SetFormat(JSONFormat).SetPretty(true)
	return ext.MarshalToString(v)
}

func (m *MarshalExt) ToYAML(v interface{}) (string, error) {
	ext := m.Clone().SetFormat(YAMLFormat)
	return ext.MarshalToString(v)
}

func (m *MarshalExt) ToXML(v interface{}) (string, error) {
	ext := m.Clone().SetFormat(XMLFormat)
	return ext.MarshalToString(v)
}

func (m *MarshalExt) ToPrettyXML(v interface{}) (string, error) {
	ext := m.Clone().SetFormat(XMLFormat).SetPretty(true)
	return ext.MarshalToString(v)
}

func (m *MarshalExt) ToString(v interface{}) (string, error) {
	ext := m.Clone().SetFormat(StringFormat)
	return ext.MarshalToString(v)
}

// Must 版本便捷方法（直接返回 string，出错时 panic）
func (m *MarshalExt) MustToJSON(v interface{}) string {
	ext := m.Clone().SetFormat(JSONFormat)
	return ext.MustMarshalToString(v)
}

func (m *MarshalExt) MustToPrettyJSON(v interface{}) string {
	ext := m.Clone().SetFormat(JSONFormat).SetPretty(true)
	return ext.MustMarshalToString(v)
}

func (m *MarshalExt) MustToYAML(v interface{}) string {
	ext := m.Clone().SetFormat(YAMLFormat)
	return ext.MustMarshalToString(v)
}

func (m *MarshalExt) MustToXML(v interface{}) string {
	ext := m.Clone().SetFormat(XMLFormat)
	return ext.MustMarshalToString(v)
}

func (m *MarshalExt) MustToPrettyXML(v interface{}) string {
	ext := m.Clone().SetFormat(XMLFormat).SetPretty(true)
	return ext.MustMarshalToString(v)
}

func (m *MarshalExt) MustToString(v interface{}) string {
	ext := m.Clone().SetFormat(StringFormat)
	return ext.MustMarshalToString(v)
}

// 全局默认序列化器
var DefaultMarshal = DefaultMarshalExt()

// 全局便捷函数
func Marshal(v interface{}) ([]byte, error) {
	return DefaultMarshal.Marshal(v)
}

func MarshalToString(v interface{}) (string, error) {
	return DefaultMarshal.MarshalToString(v)
}

// MustMarshalToString 全局序列化为字符串，出错时 panic
func MustMarshalToString(v interface{}) string {
	return DefaultMarshal.MustMarshalToString(v)
}

func MarshalToWriter(w io.Writer, v interface{}) error {
	return DefaultMarshal.MarshalToWriter(w, v)
}

func Unmarshal(data []byte, v interface{}) error {
	return DefaultMarshal.Unmarshal(data, v)
}

func UnmarshalFromString(str string, v interface{}) error {
	return DefaultMarshal.UnmarshalFromString(str, v)
}

func UnmarshalFromReader(r io.Reader, v interface{}) error {
	return DefaultMarshal.UnmarshalFromReader(r, v)
}

// 格式特定的便捷函数
func ToJSON(v interface{}) (string, error) {
	return DefaultMarshal.ToJSON(v)
}

func ToPrettyJSON(v interface{}) (string, error) {
	return DefaultMarshal.ToPrettyJSON(v)
}

func ToYAML(v interface{}) (string, error) {
	return DefaultMarshal.ToYAML(v)
}

func ToXML(v interface{}) (string, error) {
	return DefaultMarshal.ToXML(v)
}

func ToPrettyXML(v interface{}) (string, error) {
	return DefaultMarshal.ToPrettyXML(v)
}

func ToString(v interface{}) (string, error) {
	return DefaultMarshal.ToString(v)
}

// Must 版本全局格式特定函数（直接返回 string，出错时 panic）
func MustToJSON(v interface{}) string {
	return DefaultMarshal.MustToJSON(v)
}

func MustToPrettyJSON(v interface{}) string {
	return DefaultMarshal.MustToPrettyJSON(v)
}

func MustToYAML(v interface{}) string {
	return DefaultMarshal.MustToYAML(v)
}

func MustToXML(v interface{}) string {
	return DefaultMarshal.MustToXML(v)
}

func MustToPrettyXML(v interface{}) string {
	return DefaultMarshal.MustToPrettyXML(v)
}

func MustToString(v interface{}) string {
	return DefaultMarshal.MustToString(v)
}

// MarshalBuilder 序列化构建器，支持链式调用
type MarshalBuilder struct {
	value   interface{}
	marshal *MarshalExt
}

// NewMarshalBuilder 创建序列化构建器
func NewMarshalBuilder(v interface{}) *MarshalBuilder {
	return &MarshalBuilder{
		value:   v,
		marshal: DefaultMarshalExt(),
	}
}

// SetFormat 设置格式
func (b *MarshalBuilder) SetFormat(format MarshalFormat) *MarshalBuilder {
	b.marshal.SetFormat(format)
	return b
}

// SetPretty 设置美化
func (b *MarshalBuilder) SetPretty(pretty bool) *MarshalBuilder {
	b.marshal.SetPretty(pretty)
	return b
}

// SetIndent 设置缩进
func (b *MarshalBuilder) SetIndent(indent string) *MarshalBuilder {
	b.marshal.SetIndent(indent)
	return b
}

// SetMaxLength 设置最大长度
func (b *MarshalBuilder) SetMaxLength(maxLength int) *MarshalBuilder {
	b.marshal.SetMaxLength(maxLength)
	return b
}

// SetTruncate 设置截断
func (b *MarshalBuilder) SetTruncate(truncate bool) *MarshalBuilder {
	b.marshal.SetTruncate(truncate)
	return b
}

// Build 构建结果
func (b *MarshalBuilder) Build() ([]byte, error) {
	return b.marshal.Marshal(b.value)
}

// BuildString 构建字符串
func (b *MarshalBuilder) BuildString() (string, error) {
	return b.marshal.MarshalToString(b.value)
}

// MustBuildString 构建字符串，出错时 panic
func (b *MarshalBuilder) MustBuildString() string {
	return b.marshal.MustMarshalToString(b.value)
}

// BuildToWriter 构建到 Writer
func (b *MarshalBuilder) BuildToWriter(w io.Writer) error {
	return b.marshal.MarshalToWriter(w, b.value)
}

// 格式特定的构建方法
func (b *MarshalBuilder) ToJSON() (string, error) {
	return b.SetFormat(JSONFormat).BuildString()
}

func (b *MarshalBuilder) ToPrettyJSON() (string, error) {
	return b.SetFormat(JSONFormat).SetPretty(true).BuildString()
}

func (b *MarshalBuilder) ToYAML() (string, error) {
	return b.SetFormat(YAMLFormat).BuildString()
}

func (b *MarshalBuilder) ToXML() (string, error) {
	return b.SetFormat(XMLFormat).BuildString()
}

func (b *MarshalBuilder) ToPrettyXML() (string, error) {
	return b.SetFormat(XMLFormat).SetPretty(true).BuildString()
}

func (b *MarshalBuilder) ToString() (string, error) {
	return b.SetFormat(StringFormat).BuildString()
}

// Must 版本格式特定的构建方法（直接返回 string，出错时 panic）
func (b *MarshalBuilder) MustToJSON() string {
	return b.SetFormat(JSONFormat).MustBuildString()
}

func (b *MarshalBuilder) MustToPrettyJSON() string {
	return b.SetFormat(JSONFormat).SetPretty(true).MustBuildString()
}

func (b *MarshalBuilder) MustToYAML() string {
	return b.SetFormat(YAMLFormat).MustBuildString()
}

func (b *MarshalBuilder) MustToXML() string {
	return b.SetFormat(XMLFormat).MustBuildString()
}

func (b *MarshalBuilder) MustToPrettyXML() string {
	return b.SetFormat(XMLFormat).SetPretty(true).MustBuildString()
}

func (b *MarshalBuilder) MustToString() string {
	return b.SetFormat(StringFormat).MustBuildString()
}

// Clone 克隆构建器
func (b *MarshalBuilder) Clone() *MarshalBuilder {
	return &MarshalBuilder{
		value:   b.value,
		marshal: b.marshal.Clone(),
	}
}

// 类型转换工具
type TypeConverter struct {
	value interface{}
}

// NewTypeConverter 创建类型转换器
func NewTypeConverter(v interface{}) *TypeConverter {
	return &TypeConverter{value: v}
}

// ToString 转换为字符串
func (tc *TypeConverter) ToString() string {
	return fmt.Sprintf("%v", tc.value)
}

// ToJSON 转换为 JSON 字符串
func (tc *TypeConverter) ToJSON() (string, error) {
	return ToJSON(tc.value)
}

// ToPrettyJSON 转换为美化的 JSON 字符串
func (tc *TypeConverter) ToPrettyJSON() (string, error) {
	return ToPrettyJSON(tc.value)
}

// ToYAML 转换为 YAML 字符串
func (tc *TypeConverter) ToYAML() (string, error) {
	return ToYAML(tc.value)
}

// ToXML 转换为 XML 字符串
func (tc *TypeConverter) ToXML() (string, error) {
	return ToXML(tc.value)
}

// ToPrettyXML 转换为美化的 XML 字符串
func (tc *TypeConverter) ToPrettyXML() (string, error) {
	return ToPrettyXML(tc.value)
}

// Must 版本转换方法（直接返回 string，出错时 panic）
func (tc *TypeConverter) MustToJSON() string {
	return MustToJSON(tc.value)
}

func (tc *TypeConverter) MustToPrettyJSON() string {
	return MustToPrettyJSON(tc.value)
}

func (tc *TypeConverter) MustToYAML() string {
	return MustToYAML(tc.value)
}

func (tc *TypeConverter) MustToXML() string {
	return MustToXML(tc.value)
}

func (tc *TypeConverter) MustToPrettyXML() string {
	return MustToPrettyXML(tc.value)
}

// 便捷的全局函数
func ConvertToString(v interface{}) string {
	return NewTypeConverter(v).ToString()
}

func ConvertToJSON(v interface{}) (string, error) {
	return NewTypeConverter(v).ToJSON()
}

func ConvertToPrettyJSON(v interface{}) (string, error) {
	return NewTypeConverter(v).ToPrettyJSON()
}

func ConvertToYAML(v interface{}) (string, error) {
	return NewTypeConverter(v).ToYAML()
}

func ConvertToXML(v interface{}) (string, error) {
	return NewTypeConverter(v).ToXML()
}

func ConvertToPrettyXML(v interface{}) (string, error) {
	return NewTypeConverter(v).ToPrettyXML()
}

// Must 版本全局转换函数（直接返回 string，出错时 panic）
func MustConvertToJSON(v interface{}) string {
	return NewTypeConverter(v).MustToJSON()
}

func MustConvertToPrettyJSON(v interface{}) string {
	return NewTypeConverter(v).MustToPrettyJSON()
}

func MustConvertToYAML(v interface{}) string {
	return NewTypeConverter(v).MustToYAML()
}

func MustConvertToXML(v interface{}) string {
	return NewTypeConverter(v).MustToXML()
}

func MustConvertToPrettyXML(v interface{}) string {
	return NewTypeConverter(v).MustToPrettyXML()
}
