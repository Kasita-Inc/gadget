package generics

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Kasita-Inc/gadget/errors"
)

const (
	prefixLength = 4
	// Do not change these like ever.
	intPrefix     = "int:"
	int32Prefix   = "i32:"
	int64Prefix   = "i64:"
	uIntPrefix    = "uit:"
	uInt32Prefix  = "u32:"
	uInt64Prefix  = "u64:"
	stringPrefix  = "str:"
	boolPrefix    = "bln:"
	float32Prefix = "f32:"
	float64Prefix = "f64:"

	// Type converstion identifiers
	// TypeBoolean indicates that the expected value is a boolean
	TypeBoolean = "boolean"
	// TypeInteger indicates that the expected value is an integer
	TypeInteger = "integer"
	// TypeString indicates that the expected value is a string
	TypeString = "string"
	// TypeFloat indicates that the expected value is a float
	TypeFloat = "float"
	// TypeJSON indicates that the expected value is a json string
	TypeJSON = "json"
)

// ValueHolder is like a union with the underlying value stored as a serialized string
// and methods for setting and retrieving it as a string.
type ValueHolder struct {
	Value string
	Error errors.TracerError
}

func (vh *ValueHolder) clean(prefix string) (string, bool) {
	if !strings.HasPrefix(vh.Value, prefix) {
		vh.Error = errors.New("prefix %s not found in '%s'", prefix, vh.Value)
		return "", false
	}
	if len(vh.Value) == len(prefix) {
		// this value is just the prefix and is thus 'empty'
		return "", true
	}
	return vh.Value[len(prefix):], true
}

// NewValueHolder set to the value passed. This is not a raw string value from
// another value holder. Use SetRaw for that.
func NewValueHolder(value interface{}) (*ValueHolder, bool) {
	vh := &ValueHolder{}
	return vh, vh.Set(value)
}

// GetRaw value underlying this value holder. Should only be used as
// part of a marshalling routine.
func (vh *ValueHolder) GetRaw() string {
	return vh.Value
}

// GetAsString value underlying this value holder without the prefix as a string
func (vh *ValueHolder) GetAsString() string {
	if len(vh.Value) > prefixLength {
		return vh.Value[prefixLength:]
	}
	return vh.Value
}

// SetRaw value underlying this value holder. Should only be used when
// reconstituting value holder as part of unmarshalling a message.
func (vh *ValueHolder) SetRaw(value string) {
	vh.Error = nil
	vh.Value = strings.TrimSpace(value)
}

// Set the value on this value holder to the correct value by inspection and
// type assertion.
func (vh *ValueHolder) Set(value interface{}) bool {
	success := true
	vh.Error = nil
	switch v := value.(type) {
	case int:
		vh.SetInt(value.(int))
	case int32:
		vh.SetInt32(value.(int32))
	case int64:
		vh.SetInt64(value.(int64))
	case uint:
		vh.SetUInt(value.(uint))
	case uint32:
		vh.SetUInt32(value.(uint32))
	case uint64:
		vh.SetUInt64(value.(uint64))
	case float32:
		vh.SetFloat32(value.(float32))
	case float64:
		vh.SetFloat64(value.(float64))
	case string:
		vh.SetString(value.(string))
	case bool:
		vh.SetBool(value.(bool))
	default:
		vh.Error = NewUnsupportedTypeError(v)
		success = false
	}
	return success
}

// ParseSet will set the ValueHolder with the string, if it's not in a proper ValueHolder format it will attempt to
// parse the string into the target type
func (vh *ValueHolder) ParseSet(value string, targetType string) error {
	vh.SetRaw(value)
	_, ok := vh.Get()
	if !ok {
		var obj interface{}
		var err error
		switch targetType {
		case TypeBoolean:
			obj, err = strconv.ParseBool(value)
		case TypeInteger:
			obj, err = strconv.ParseInt(value, 10, 64)
		case TypeString:
			obj = value
		case TypeFloat:
			obj, err = strconv.ParseFloat(value, 32)
		case TypeJSON:
			obj = value
		default:
			err = fmt.Errorf("unsupported function field type %s", targetType)
		}
		if nil != err {
			vh.SetRaw("")
			return err
		}
		vh.Set(obj)
	}
	return vh.Error
}

// Get the actual type stored in this ValueHolder as an anonymized type.
func (vh *ValueHolder) Get() (interface{}, bool) {
	vh.Error = nil
	success := true
	if len(vh.Value) < 4 {
		vh.Error = NewCorruptValueError(vh.Value)
		return nil, false
	}
	var v interface{}
	switch vh.Value[0:prefixLength] {
	case intPrefix:
		v, success = vh.GetInt()
	case int32Prefix:
		v, success = vh.GetInt32()
	case int64Prefix:
		v, success = vh.GetInt64()
	case uIntPrefix:
		v, success = vh.GetUint()
	case uInt32Prefix:
		v, success = vh.GetUInt32()
	case uInt64Prefix:
		v, success = vh.GetUInt64()
	case float32Prefix:
		v, success = vh.GetFloat32()
	case float64Prefix:
		v, success = vh.GetFloat64()
	case stringPrefix:
		v, success = vh.GetString()
	case boolPrefix:
		v, success = vh.GetBool()
	default:
		vh.Error = NewCorruptValueError(vh.Value)
		success = false
	}
	return v, success
}

func (vh *ValueHolder) getInt(prefix string, size int) (int64, bool) {
	vh.Error = nil
	clean, ok := vh.clean(prefix)
	if !ok {
		return 0, false
	}
	i, err := strconv.ParseInt(clean, 10, size)
	if nil != err {
		vh.Error = errors.Wrap(err)
		return 0, false
	}
	return i, true
}

// GetInt from this value holder.
func (vh *ValueHolder) GetInt() (int, bool) {
	value := 0
	if i, ok := vh.getInt(intPrefix, 64); ok {
		value = int(i)
	} else if i, ok := vh.getInt(int64Prefix, 64); ok {
		value = int(i)
	} else if i, ok := vh.getInt(int32Prefix, 32); ok {
		value = int(i)
	} else {
		return 0, false
	}
	return value, true
}

// GetInt32 value from this value holder
func (vh *ValueHolder) GetInt32() (int32, bool) {
	i64, ok := vh.getInt(int32Prefix, 32)
	return int32(i64), ok
}

// SetInt into this value holder
func (vh *ValueHolder) SetInt(i int) {
	b := []byte(intPrefix)
	vh.Value = string(strconv.AppendInt(b, int64(i), 10))
}

// SetInt32 into this value holder.
func (vh *ValueHolder) SetInt32(i32 int32) {
	b := []byte(int32Prefix)
	vh.Value = string(strconv.AppendInt(b, int64(i32), 10))
}

// GetInt64 value from this value holder
func (vh *ValueHolder) GetInt64() (int64, bool) {
	return vh.getInt(int64Prefix, 64)
}

// SetInt64 into this value holder.
func (vh *ValueHolder) SetInt64(i64 int64) {
	b := []byte(int64Prefix)
	vh.Value = string(strconv.AppendInt(b, i64, 10))
}

func (vh *ValueHolder) getUInt(prefix string, size int) (uint64, bool) {
	vh.Error = nil
	clean, ok := vh.clean(prefix)
	if !ok {
		return 0, false
	}
	i, err := strconv.ParseUint(clean, 10, size)
	if nil != err {
		vh.Error = errors.Wrap(err)
		return 0, false
	}
	return i, true
}

// GetUint value form this value holder
func (vh *ValueHolder) GetUint() (uint, bool) {
	var value uint
	if i, ok := vh.getUInt(uIntPrefix, 64); ok {
		value = uint(i)
	} else if i, ok := vh.getUInt(uInt32Prefix, 32); ok {
		value = uint(i)
	} else if i, ok := vh.getUInt(uInt64Prefix, 64); ok {
		value = uint(i)
	} else {
		return 0, false
	}
	return value, true
}

// SetUInt into this value holder.
func (vh *ValueHolder) SetUInt(u uint) {
	b := []byte(uIntPrefix)
	vh.Value = string(strconv.AppendUint(b, uint64(u), 10))
}

// GetUInt32 value from this value holder
func (vh *ValueHolder) GetUInt32() (uint32, bool) {
	u64, ok := vh.getUInt(uInt32Prefix, 32)
	return uint32(u64), ok
}

// SetUInt32 into this value holder.
func (vh *ValueHolder) SetUInt32(u32 uint32) {
	b := []byte(uInt32Prefix)
	vh.Value = string(strconv.AppendUint(b, uint64(u32), 10))
}

// GetUInt64 value from this value holder
func (vh *ValueHolder) GetUInt64() (uint64, bool) {
	return vh.getUInt(uInt64Prefix, 64)
}

// SetUInt64 into this value holder.
func (vh *ValueHolder) SetUInt64(u64 uint64) {
	b := []byte(uInt64Prefix)
	vh.Value = string(strconv.AppendUint(b, u64, 10))
}

// GetString value from this value holder
func (vh *ValueHolder) GetString() (string, bool) {
	return vh.clean(stringPrefix)
}

// SetString into this value holder
func (vh *ValueHolder) SetString(s string) {
	vh.Value = stringPrefix + s
}

// GetBool value from this value holder
func (vh *ValueHolder) GetBool() (bool, bool) {
	vh.Error = nil
	clean, ok := vh.clean(boolPrefix)
	if !ok {
		vh.Error = errors.New("bad prefix, expected '%s' but raw value was '%s'", boolPrefix, vh.GetRaw())
		return false, false
	}
	b, err := strconv.ParseBool(clean)
	if nil != err {
		vh.Error = errors.Wrap(err)
		return false, false
	}
	return b, true
}

// SetBool into this value holder
func (vh *ValueHolder) SetBool(v bool) {
	b := []byte(boolPrefix)
	vh.Value = string(strconv.AppendBool(b, v))
}

func (vh *ValueHolder) getFloat(prefix string, size int) (float64, bool) {
	vh.Error = nil
	clean, ok := vh.clean(prefix)
	if !ok {
		return 0, false
	}
	i, err := strconv.ParseFloat(clean, size)
	if nil != err {
		vh.Error = errors.Wrap(err)
		return 0, false
	}
	return i, true
}

// SetFloat32 into this value holder
func (vh *ValueHolder) SetFloat32(f32 float32) {
	b := []byte(float32Prefix)
	vh.Value = string(strconv.AppendFloat(b, float64(f32), 'f', -1, 32))
}

// GetFloat32 value from this value holder
func (vh *ValueHolder) GetFloat32() (float32, bool) {
	f, ok := vh.getFloat(float32Prefix, 32)
	return float32(f), ok
}

// SetFloat64 into this value holder
func (vh *ValueHolder) SetFloat64(f64 float64) {
	b := []byte(float64Prefix)
	vh.Value = string(strconv.AppendFloat(b, f64, 'f', -1, 64))
}

// GetFloat64 value from this value holder
func (vh *ValueHolder) GetFloat64() (float64, bool) {
	return vh.getFloat(float64Prefix, 64)
}

func init() {
	// guards
	if reflect.TypeOf(int(0)).Size() > 64 {
		panic("bit size of 'int' must be less than or equal to 64 to use valueholder")
	}

	if reflect.TypeOf(uint(0)).Size() > 64 {
		panic("bit size of 'uint' must be less than or equal to 64 to use valueholder")
	}
}
