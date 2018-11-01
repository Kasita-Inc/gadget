package generics

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/errors"
)

type fields struct {
	value string
	Error errors.TracerError
}

func TestValueHolder_clean(t *testing.T) {
	type args struct {
		prefix string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			name:   "empty value",
			fields: fields{value: ""},
			args:   args{prefix: "prefix"},
			want:   "",
			want1:  false,
		},
		{
			name:   "value does not contain prefix",
			fields: fields{value: "asdf"},
			args:   args{prefix: "prefix"},
			want:   "",
			want1:  false,
		},
		{
			name:   "values contains prefix exactly",
			fields: fields{value: "prefix"},
			args:   args{prefix: "prefix"},
			want:   "",
			want1:  true,
		},
		{
			name:   "works correctly",
			fields: fields{value: "prefix1234"},
			args:   args{prefix: "prefix"},
			want:   "1234",
			want1:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			got, got1 := vh.clean(tt.args.prefix)
			if got != tt.want {
				t.Errorf("ValueHolder.clean() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValueHolder.clean() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewValueHolder(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  *ValueHolder
		want1 bool
	}{
		{
			name:  "string",
			args:  args{value: "string"},
			want:  &ValueHolder{Value: stringPrefix + "string"},
			want1: true,
		},
		{
			name:  "empty string",
			args:  args{value: ""},
			want:  &ValueHolder{Value: stringPrefix},
			want1: true,
		},
		{
			name:  "int",
			args:  args{value: 10},
			want:  &ValueHolder{Value: intPrefix + "10"},
			want1: true,
		},
		{
			name:  "int32",
			args:  args{value: int32(10)},
			want:  &ValueHolder{Value: int32Prefix + "10"},
			want1: true,
		},
		{
			name:  "int64",
			args:  args{value: int64(10)},
			want:  &ValueHolder{Value: int64Prefix + "10"},
			want1: true,
		},
		{
			name:  "uint",
			args:  args{value: uint(10)},
			want:  &ValueHolder{Value: uIntPrefix + "10"},
			want1: true,
		},
		{
			name:  "uint32",
			args:  args{value: uint32(10)},
			want:  &ValueHolder{Value: uInt32Prefix + "10"},
			want1: true,
		},
		{
			name:  "uint64",
			args:  args{value: uint64(10)},
			want:  &ValueHolder{Value: uInt64Prefix + "10"},
			want1: true,
		},
		{
			name:  "float32",
			args:  args{value: float32(10.01)},
			want:  &ValueHolder{Value: float32Prefix + "10.01"},
			want1: true,
		},
		{
			name:  "float64",
			args:  args{value: float64(10.1234)},
			want:  &ValueHolder{Value: float64Prefix + "10.1234"},
			want1: true,
		},
		{
			name:  "bool",
			args:  args{value: true},
			want:  &ValueHolder{Value: boolPrefix + "true"},
			want1: true,
		},
		{
			name:  "error",
			args:  args{value: nil},
			want:  &ValueHolder{Value: "", Error: NewUnsupportedTypeError(nil)},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := NewValueHolder(tt.args.value)
			if !reflect.DeepEqual(got.Value, tt.want.Value) {
				t.Errorf("NewValueHolder().value got = %v, want %v", got.Value, tt.want.Value)
			}
			if !(reflect.DeepEqual(got.Error, tt.want.Error) || reflect.DeepEqual(got.Error.Error(), tt.want.Error.Error())) {
				t.Errorf("NewValueHolder().Error got = %v, want %v", got.Error, tt.want.Error)
			}
			if got1 != tt.want1 {
				t.Errorf("NewValueHolder() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestValueHolder_GetRaw(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{value: "asdf", Error: nil},
			want:   "asdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			if got := vh.GetRaw(); got != tt.want {
				t.Errorf("ValueHolder.GetRaw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueHolder_SetRaw(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			fields: fields{value: "awef"},
			args:   args{value: "awef"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			vh.SetRaw(tt.args.value)
		})
	}
}

func TestValueHolder_Set(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "int",
			fields: fields{value: intPrefix + "10", Error: nil},
			args:   args{value: int(10)},
			want:   true,
		},
		{
			name:   "int32",
			fields: fields{value: int32Prefix + "10", Error: nil},
			args:   args{value: int32(10)},
			want:   true,
		},
		{
			name:   "int64",
			fields: fields{value: int64Prefix + "10", Error: nil},
			args:   args{value: int64(10)},
			want:   true,
		},
		{
			name:   "uInt",
			fields: fields{value: uIntPrefix + "10", Error: nil},
			args:   args{value: uint(10)},
			want:   true,
		},
		{
			name:   "uInt32",
			fields: fields{value: uInt32Prefix + "10", Error: nil},
			args:   args{value: uint32(10)},
			want:   true,
		},
		{
			name:   "uInt64",
			fields: fields{value: uInt64Prefix + "10", Error: nil},
			args:   args{value: uint64(10)},
			want:   true,
		},
		{
			name:   "float32",
			fields: fields{value: float32Prefix + "10.1234", Error: nil},
			args:   args{value: float32(10.1234)},
			want:   true,
		},
		{
			name:   "float64",
			fields: fields{value: float64Prefix + "10.1234", Error: nil},
			args:   args{value: float64(10.1234)},
			want:   true,
		},
		{
			name:   "string",
			fields: fields{value: stringPrefix + "asdf", Error: nil},
			args:   args{value: "asdf"},
			want:   true,
		},
		{
			name:   "bool",
			fields: fields{value: boolPrefix + "true", Error: nil},
			args:   args{value: true},
			want:   true,
		},
		{
			name:   "error",
			fields: fields{value: "monkeys", Error: NewCorruptValueError("monkeys")},
			args:   args{value: nil},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			if got := vh.Set(tt.args.value); got != tt.want {
				t.Errorf("ValueHolder.Set() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueHolder_Get(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   interface{}
		want1  bool
	}{
		{
			name:   "int",
			fields: fields{value: intPrefix + "10", Error: nil},
			want:   int(10),
			want1:  true,
		},
		{
			name:   "int32",
			fields: fields{value: int32Prefix + "10", Error: nil},
			want:   int32(10),
			want1:  true,
		},
		{
			name:   "int64",
			fields: fields{value: int64Prefix + "10", Error: nil},
			want:   int64(10),
			want1:  true,
		},
		{
			name:   "uInt",
			fields: fields{value: uIntPrefix + "10", Error: nil},
			want:   uint(10),
			want1:  true,
		},
		{
			name:   "uInt32",
			fields: fields{value: uInt32Prefix + "10", Error: nil},
			want:   uint32(10),
			want1:  true,
		},
		{
			name:   "uInt64",
			fields: fields{value: uInt64Prefix + "10", Error: nil},
			want:   uint64(10),
			want1:  true,
		},
		{
			name:   "float32",
			fields: fields{value: float32Prefix + "10.1234", Error: nil},
			want:   float32(10.1234),
			want1:  true,
		},
		{
			name:   "float64",
			fields: fields{value: float64Prefix + "10.1234", Error: nil},
			want:   float64(10.1234),
			want1:  true,
		},
		{
			name:   "string",
			fields: fields{value: stringPrefix + "asdf", Error: nil},
			want:   "asdf",
			want1:  true,
		},
		{
			name:   "bool",
			fields: fields{value: boolPrefix + "true", Error: nil},
			want:   true,
			want1:  true,
		},
		{
			name:   "error",
			fields: fields{value: "monkeys"},
			want:   nil,
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			got, got1 := vh.Get()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValueHolder.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValueHolder.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestValueHolder_getInt(t *testing.T) {
	type args struct {
		prefix string
		size   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  bool
	}{
		{
			name:   "fail no prefix",
			fields: fields{value: "10"},
			want:   0,
			want1:  false,
		},
		{
			name:   "fail bad parse",
			fields: fields{value: intPrefix + "asdf"},
			want:   0,
			want1:  false,
		},
		{
			name:   "int prefix",
			fields: fields{value: intPrefix + "10"},
			want:   10,
			want1:  true,
		},
		{
			name:   "int32 prefix",
			fields: fields{value: int32Prefix + "10"},
			want:   10,
			want1:  true,
		},
		{
			name:   "int64 prefix",
			fields: fields{value: int64Prefix + "10"},
			want:   10,
			want1:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			got, got1 := vh.GetInt()
			if got != tt.want {
				t.Errorf("ValueHolder.getInt() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValueHolder.getInt() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestValueHolder_getUInt(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   uint
		want1  bool
	}{
		{
			name:   "fail no prefix",
			fields: fields{value: "10"},
			want:   0,
			want1:  false,
		},
		{
			name:   "fail bad parse",
			fields: fields{value: uIntPrefix + "asdf"},
			want:   0,
			want1:  false,
		},
		{
			name:   "works",
			fields: fields{value: uIntPrefix + "10"},
			want:   10,
			want1:  true,
		},
		{
			name:   "uInt32",
			fields: fields{value: uInt32Prefix + "10"},
			want:   10,
			want1:  true,
		},
		{
			name:   "uInt64",
			fields: fields{value: uInt64Prefix + "10"},
			want:   10,
			want1:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			got, got1 := vh.GetUint()
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func TestValueHolder_GetBool(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
		want1  bool
	}{
		{
			name:   "works",
			fields: fields{value: boolPrefix + "true"},
			want:   true,
			want1:  true,
		},
		{
			name:   "bad prefix",
			fields: fields{value: uIntPrefix + "true"},
			want:   false,
			want1:  false,
		},
		{
			name:   "bad parse",
			fields: fields{value: boolPrefix + "tralse"},
			want:   false,
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			got, got1 := vh.GetBool()
			if got != tt.want {
				t.Errorf("ValueHolder.GetBool() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValueHolder.GetBool() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestValueHolder_getFloat(t *testing.T) {
	type args struct {
		prefix string
		size   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
		want1  bool
	}{
		{
			name:   "fail no prefix",
			fields: fields{value: "10"},
			args:   args{prefix: float32Prefix, size: 32},
			want:   0,
			want1:  false,
		},
		{
			name:   "fail bad parse",
			fields: fields{value: float32Prefix + "asdf"},
			args:   args{prefix: float32Prefix, size: 32},
			want:   0,
			want1:  false,
		},
		{
			name:   "works",
			fields: fields{value: float64Prefix + "10.1"},
			args:   args{prefix: float64Prefix, size: 64},
			want:   10.1,
			want1:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh := &ValueHolder{
				Value: tt.fields.value,
				Error: tt.fields.Error,
			}
			got, got1 := vh.getFloat(tt.args.prefix, tt.args.size)
			if got != tt.want {
				t.Errorf("ValueHolder.getFloat() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValueHolder.getFloat() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestValueHolder_ParseSet(t *testing.T) {
	type args struct {
		value     string
		fieldType string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "string",
			args: args{
				value:     "foo",
				fieldType: TypeString,
			},
			want:    "str:foo",
			wantErr: false,
		},
		{
			name: "bool",
			args: args{
				value:     "true",
				fieldType: TypeBoolean,
			},
			want:    "bln:true",
			wantErr: false,
		},
		{
			name: "int",
			args: args{
				value:     "12",
				fieldType: TypeInteger,
			},
			want:    "i64:12",
			wantErr: false,
		},
		{
			name: "JSON",
			args: args{
				value:     `{"h":0,"i":1}`,
				fieldType: TypeJSON,
			},
			want:    `str:{"h":0,"i":1}`,
			wantErr: false,
		},
		{
			name: "float",
			args: args{
				value:     "10",
				fieldType: TypeFloat,
			},
			want:    "f64:10",
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				value:     "10",
				fieldType: "unsupported",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh, _ := NewValueHolder(nil)
			err := vh.ParseSet(tt.args.value, tt.args.fieldType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatabaseValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if vh.GetRaw() != tt.want {
				t.Errorf("GetDatabaseValue() = %v, want %v", vh.GetRaw(), tt.want)
			}
		})
	}
}
