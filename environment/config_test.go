package environment

import (
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/generator"
)

type specification struct {
	StringField         string `env:"STRING_FIELD" s3:"bar,foo"`
	IntField            int    `env:"INT_FIELD,junk" s3:"foo,bar"`
	OptionalField       string `env:"OPTIONAL_FIELD,optional,junk"`
	NotEnvironmentField string
}

type unsupportedTypeSpecification struct {
	BoolField bool `env:"BOOL_FIELD" s3:"invalid,type"`
}

func TestPush(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()

	spec := &specification{
		StringField:   generator.String(20),
		IntField:      10,
		OptionalField: generator.String(30),
	}

	assert.NoError(Push(spec))
	assert.Equal(os.Getenv("STRING_FIELD"), spec.StringField)
	assert.Equal(os.Getenv("INT_FIELD"), "10")
	assert.Equal(os.Getenv("OPTIONAL_FIELD"), spec.OptionalField)
}

func TestValidConfig(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	expectedIntField := 42
	os.Setenv("INT_FIELD", strconv.Itoa(expectedIntField))

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := &specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config)

	assert.NoError(err)
	assert.Equal(expectedStringField, config.StringField)
	assert.Equal(expectedIntField, config.IntField)
	assert.Equal("", config.OptionalField)
	assert.Equal(expectedNotEnvironmentField, config.NotEnvironmentField)
}

func TestAWSLookupDuringProcess(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()
	os.Setenv("STRING_FIELD", awsMetaService+"amazon.com")
	expectedIntField := 42
	os.Setenv("INT_FIELD", strconv.Itoa(expectedIntField))
	config := &specification{OptionalField: awsMetaService + "amazon.com"}
	err := Process(config)
	awsLookupReturn := "<html>\r\n<body>\r\n</body>\r\n</html>\r\n"

	assert.NoError(err)
	assert.Equal(&specification{StringField: awsLookupReturn, IntField: expectedIntField, OptionalField: awsLookupReturn}, config)
}

func TestProcessNonPointerFails(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	expectedIntField := 42
	os.Setenv("INT_FIELD", strconv.Itoa(expectedIntField))

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config)

	assert.EqualError(err, NewInvalidSpecificationError().Error())
	assert.Equal(specification{NotEnvironmentField: expectedNotEnvironmentField}, config)
}

func TestMissingEnviroment(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := &specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config)

	assert.EqualError(err, MissingEnvironmentVariableError{Tag: "INT_FIELD", Field: "IntField"}.Error())
	assert.Equal(expectedStringField, config.StringField)
	assert.Equal(0, config.IntField)
	assert.Equal(expectedNotEnvironmentField, config.NotEnvironmentField)
}

func TestNotImplementedType(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()

	os.Setenv("BOOL_FIELD", "true")

	config := &unsupportedTypeSpecification{}
	err := Process(config)

	assert.EqualError(err, UnsupportedDataTypeError{Type: reflect.Bool, Field: "BoolField"}.Error())
	assert.Equal(&unsupportedTypeSpecification{}, config)
}

func TestInvalidConfigValue(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	os.Setenv("INT_FIELD", "j")

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := &specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config)

	assert.Error(err)
	assert.Equal("strconv.Atoi: parsing \"j\": invalid syntax while converting INT_FIELD", err.Error())
	assert.Equal(expectedStringField, config.StringField)
	assert.Equal(0, config.IntField)
	assert.Equal("", config.OptionalField)
	assert.Equal(expectedNotEnvironmentField, config.NotEnvironmentField)
}

func TestNonStructProcessed(t *testing.T) {
	assert := assert.New(t)
	os.Clearenv()

	config := "42"
	err := Process(&config)

	assert.EqualError(err, NewInvalidSpecificationError().Error())
}
