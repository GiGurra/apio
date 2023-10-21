package apio

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type pathFieldSetter = func(target reflect.Value, from string) error

type queryFieldSetter = func(target reflect.Value, from *string) error
type headerFieldSetter = func(target reflect.Value, from *string) error

func getFromStringPathFieldSetter(field reflect.StructField) pathFieldSetter {
	parseFn, err := getStringParsePtrFn(field.Type)
	if err != nil {
		panic(fmt.Errorf("failed to get parse function for field '%s': %w", field.Name, err))
	}

	return func(target reflect.Value, from string) error {
		parsedPtr, err := parseFn(from)
		if err != nil {
			return fmt.Errorf("failed to parse '%s' into field %s [%t]: %w", from, field.Name, field.Type, err)
		}
		target.Set(reflect.ValueOf(parsedPtr).Elem())
		return nil
	}
}

func getFromStringQueryFieldSetter(field reflect.StructField) queryFieldSetter {
	parseFn, err := getStringParsePtrFn(field.Type)
	if err != nil {
		panic(fmt.Errorf("failed to get parse function for field '%s': %w", field.Name, err))
	}

	return func(target reflect.Value, from *string) error {

		if from == nil {
			// Check that target is a pointer (=optional)
			if target.Kind() != reflect.Ptr {
				return fmt.Errorf("missing required parameter '%s'", field.Name)
			} else {
				// Leave the target at nil/zero/unset
				return nil
			}
		}

		parsedPtr, err := parseFn(*from)
		if err != nil {
			return fmt.Errorf("failed to parse '%s' into field %s [%t]: %w", *from, field.Name, field.Type, err)
		}
		if target.Kind() == reflect.Ptr {
			target.Set(reflect.ValueOf(parsedPtr))
		} else {
			target.Set(reflect.ValueOf(parsedPtr).Elem())
		}
		return nil
	}
}

func getFromStringHeaderFieldSetter(field reflect.StructField, name string) queryFieldSetter {
	parseFn, err := getStringParsePtrFn(field.Type)
	if err != nil {
		panic(fmt.Errorf("failed to get parse function for field '%s': %w", name, err))
	}

	return func(target reflect.Value, from *string) error {

		if from == nil {
			// Check that target is a pointer (=optional)
			if target.Kind() != reflect.Ptr {
				return fmt.Errorf("missing required header parameter '%s'", name)
			} else {
				// Leave the target at nil/zero/unset
				return nil
			}
		}

		parsedPtr, err := parseFn(*from)
		if err != nil {
			return fmt.Errorf("failed to parse '%s' into field %s [%t]: %w", *from, name, field.Type, err)
		}
		if target.Kind() == reflect.Ptr {
			target.Set(reflect.ValueOf(parsedPtr))
		} else {
			target.Set(reflect.ValueOf(parsedPtr).Elem())
		}
		return nil
	}
}

func getStringParsePtrFn(tpe reflect.Type) (func(string) (any, error), error) {

	// This is super silly. And we should probably optimize this a bit ;).
	// We just try to json deserialize first quoted, then unquoted.
	// YES this is crappy, but, we can optimize later. Moving on...

	if tpe.Kind() == reflect.Ptr {
		tpe = tpe.Elem()
	}

	return func(from string) (interface{}, error) {

		// First quoted (pretty silly, yes)
		fromQuoted := fmt.Sprintf("\"%s\"", from)

		res1 := reflect.New(tpe).Interface()
		err1 := json.Unmarshal([]byte(fromQuoted), &res1)
		if err1 == nil {
			return res1, nil
		}

		// Ok quoting didn't work, just ty it raw
		res2 := reflect.New(tpe).Interface()
		err2 := json.Unmarshal([]byte(from), &res2)
		if err2 != nil {
			// return the err1 here, because most likely it is more useful
			return nil, fmt.Errorf("failed to parse '%s' into type %t: err1=%w, err2=%w", from, tpe, err1, err2)
		}

		return res2, nil
	}, nil
}
