package hata

import (
	"fmt"

	"reflect"
	"regexp"

	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Parser struct {
	values   interface{}
	ref      reflect.Value
	fieldMap map[string]reflect.Value

	stopWords sort.StringSlice
}

type Context struct {
}

func NewParser(defaultStruct interface{}, stopWords []string) (*Parser, error) {
	inputError := fmt.Errorf("The given value is not a struct pointer.")

	refIf := reflect.ValueOf(defaultStruct)
	if !refIf.IsValid() || refIf.Kind() != reflect.Ptr {
		return nil, inputError
	}

	ref := refIf.Elem()
	if ref.Kind() != reflect.Struct {
		return nil, inputError
	}

	fieldMap := make(map[string]reflect.Value)
	numField := ref.Type().NumField()
	for i := 0; i < numField; i++ {
		field := ref.Type().Field(i)

		// skip if the kind is unsupported
		/*
			switch field.Kind() {
			case reflect.Int:
			case reflect.Int8:
			case reflect.Int16:
			case reflect.Int32:
			case reflect.Int64:
			case reflect.Uint:
			case reflect.Uint8:
			case reflect.Uint16:
			case reflect.Uint32:
			case reflect.Uint64:
			case reflect.Uintptr:
			case reflect.Float64:
			case reflect.Map:
			case reflect.Ptr:
			case reflect.Slice:
			case reflect.String:
			default:
				continue
			}
		*/
		argName := ToArgumentName(field.Name)
		if _, found := fieldMap[argName]; found {
			return nil, fmt.Errorf("Duplicated argument name found: %s", argName)
		}
		fieldMap[argName] = ref.Field(i)
		short := field.Tag.Get("short")
		if short != "" {
			if _, found := fieldMap[short]; found {
				return nil, fmt.Errorf("Duplicated short argument name found: %s", short)
			}
			fieldMap[short] = ref.Field(i)
		}

	}

	sortSlice := sort.StringSlice(stopWords)
	sortSlice.Sort()

	return &Parser{
		values:    defaultStruct,
		ref:       ref,
		fieldMap:  fieldMap,
		stopWords: sortSlice,
	}, nil
}

var (
	LongFlagWithValue            = regexp.MustCompile("^--(.+?)=(.+)$")
	LongFlag                     = regexp.MustCompile("^--(.+)$")
	ShortFlagWithValueOrCombined = regexp.MustCompile("^-(.)(.+)$")
	ShortFlag                    = regexp.MustCompile("^-(.)$")
	EndOfFlags                   = regexp.MustCompile("^--")
)

func (parser *Parser) Parse(arguments []string) (remaining []string, errs []error) {
	inputMap, remaining, mapErrors := parser.MapInput(arguments)
	scanErrors := parser.Scan(inputMap)

	errs = append(mapErrors, scanErrors...)
	return
}

func (parser *Parser) MapInput(arguments []string) (map[string][]*string, []string, []error) {
	errors := []error{}
	inputMap := make(map[string][]*string)

	it := argumentIterator{list: arguments, current: 0}
	for {
		argument, ok := it.next()
		if !ok {
			break
		}

		if matches := LongFlagWithValue.FindStringSubmatch(argument); len(matches) > 0 {
			putMap(inputMap, matches[1], &matches[2])
			continue
		}

		if matches := LongFlag.FindStringSubmatch(argument); len(matches) > 0 {
			v, ok := it.nextValue()
			value := &v
			if !ok {
				value = nil
			}
			putMap(inputMap, matches[1], value)
			continue
		}

		if matches := ShortFlagWithValueOrCombined.FindStringSubmatch(argument); len(matches) > 0 {
			name := matches[1]
			valueRequired, err := parser.CheckIfValueRequired(name)
			if valueRequired || err != nil { // don't care the existence here
				putMap(inputMap, name, &matches[2])
			} else {
				value := ""
				putMap(inputMap, name, &value)
				for _, c := range matches[2] {
					localName := string(c)
					cValueRequired, _ := parser.CheckIfValueRequired(localName)
					if cValueRequired {
						errors = append(errors, fmt.Errorf("Argument `%s` is not a binary flag, but found in a combined flag.", localName))
						continue
					}
					cValue := ""
					putMap(inputMap, localName, &cValue)
				}
			}
			continue
		}

		if matches := ShortFlag.FindStringSubmatch(argument); len(matches) > 0 {
			v, ok := it.nextValue()
			value := &v
			if !ok {
				value = nil
			}
			putMap(inputMap, matches[1], value)
			continue
		}

		// non flags
		it.back()
		break
	}

	return inputMap, it.remaining(), errors
}

func (parser *Parser) Scan(argumentMap map[string][]*string) []error {
	errors := []error{}

	for name, values := range argumentMap {
		_, ok := parser.fieldMap[name]
		if !ok {
			errors = append(errors, fmt.Errorf("Argument `%s` is an unknown argument", name))
			continue
		}

		multiple, _ := parser.CheckIfAcceptMultiple(name)
		if !multiple {
			if len(argumentMap[name]) > 1 {
				errors = append(errors, fmt.Errorf("Argument `%s` can not accept multiple values", name))
				continue
			}
		}

		valueRequired, _ := parser.CheckIfValueRequired(name)
		if valueRequired {
			localErrors := []error{}
			for _, value := range values {
				if value == nil {
					localErrors = append(localErrors, fmt.Errorf("Argument `%s` requires a value, but not provided", name))
				}
			}
			if len(localErrors) > 0 {
				errors = append(errors, localErrors...)
				continue
			}
		}

		for _, value := range values {
			err := parser.UpdateValue(name, *value)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

func (parser *Parser) CheckIfAcceptMultiple(argName string) (bool, error) {
	field, ok := parser.fieldMap[argName]

	if !ok {
		return false, fmt.Errorf("Field %s is not found", argName)
	}

	switch field.Kind() {
	case reflect.Array:
		return true, nil
	case reflect.Map:
		return true, nil
	case reflect.Slice:
		return true, nil
	default:
		return false, nil
	}
}

func (parser *Parser) CheckIfValueRequired(argName string) (bool, error) {
	field, ok := parser.fieldMap[argName]

	if !ok {
		return false, fmt.Errorf("Field %s is not found", argName)
	}

	switch field.Kind() {
	case reflect.Bool:
		return false, nil
	default:
		return true, nil
	}
}

func (parser *Parser) UpdateValue(argName string, value string) error {
	field, ok := parser.fieldMap[argName]

	if !ok {
		return fmt.Errorf("Field %s is not found", argName)
	}

	err := fillValue(field, value)
	if err != nil {
		return fmt.Errorf("Failed to parse `%s`: %s", argName, err.Error())
	}

	return nil
}

func fillValue(field reflect.Value, value string) error {
	// do not use .SetInt or .SetUint with .ParseX with 64
	// they cannot detect overflows
	switch field.Kind() {
	case reflect.Bool:
		field.Set(reflect.ValueOf(true))

	case reflect.Int:
		v, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(int(v)))
	case reflect.Int8:
		v, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(int8(v)))
	case reflect.Int16:
		v, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(int16(v)))
	case reflect.Int32:
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(int32(v)))
	case reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(int64(v)))

	case reflect.Uint:
		v, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(uint(v)))
	case reflect.Uint8:
		v, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(uint8(v)))
	case reflect.Uint16:
		v, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(uint16(v)))
	case reflect.Uint32:
		v, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(uint32(v)))
	case reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(uint64(v)))
	case reflect.Uintptr:
		v, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(uintptr(v)))

	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(float32(v)))
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return numericError(err)
		}
		field.Set(reflect.ValueOf(float64(v)))

	case reflect.Map:
		if field.IsNil() {
			field.Set(
				reflect.MakeMap(
					reflect.MapOf(
						field.Type().Key(),
						field.Type().Elem(),
					),
				),
			)
		}
		k := reflect.New(field.Type().Key())
		v := reflect.New(field.Type().Elem())
		parts := strings.Split(value, ":")
		err := fillValue(k.Elem(), parts[0])
		if err != nil {
			return err
		}
		err = fillValue(v.Elem(), parts[1])
		if err != nil {
			return err
		}
		field.SetMapIndex(k.Elem(), v.Elem())

	case reflect.Ptr:
		v := reflect.New(field.Type().Elem())
		err := fillValue(v.Elem(), value)
		if err != nil {
			return err
		}
		field.Set(v)

	case reflect.Slice:
		if field.IsNil() {
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		}
		v := reflect.New(field.Type().Elem())
		err := fillValue(v.Elem(), value)
		if err != nil {
			return err
		}
		field.Set(reflect.Append(field, v.Elem()))

	case reflect.String:
		field.SetString(value)

	default:
		return fmt.Errorf("Unsupported kind: %s.", field.Kind())
	}
	return nil
}

func numericError(err error) error {
	e := err.(*strconv.NumError)
	if e.Err == strconv.ErrRange {
		return fmt.Errorf("Given value is out of range")
	} else if e.Err == strconv.ErrSyntax {
		return fmt.Errorf("Non numeric value is provided to numeric parameter")
	} else {
		return fmt.Errorf("Numeric parse error")
	}
}

func isFlag(argument string) bool {
	if LongFlagWithValue.MatchString(argument) ||
		LongFlag.MatchString(argument) ||
		ShortFlagWithValueOrCombined.MatchString(argument) ||
		ShortFlag.MatchString(argument) {
		return true
	}

	return false
}

func putMap(to map[string][]*string, key string, value *string) {
	_, ok := to[key]
	if !ok {
		to[key] = []*string{}
	}
	to[key] = append(to[key], value)
}

var CamelCaseSplitter = regexp.MustCompile("(^[a-z]|[A-Z])[0-9a-z_]*")

func ToArgumentName(str string) string {
	return strings.ToLower(strings.Join(SplitCamelCase(str), "-"))
}

func SplitCamelCase(str string) []string {
	result := make([]string, 0)
	parts := CamelCaseSplitter.FindAllStringSubmatch(str, -1)
	for i, part := range parts {
		if (len(part[0]) == 1 || allDigit(part[0][1:])) && i > 0 && (len(parts[i-1][0]) == 1 || allDigit(parts[i-1][0][1:])) {
			result[len(result)-1] += part[0]
		} else {
			result = append(result, part[0])
		}
	}
	return result
}

func allDigit(str string) bool {
	for _, char := range str {
		if !unicode.IsNumber(char) {
			return false
		}
	}
	return true
}
