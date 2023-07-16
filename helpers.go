package filecache

import (
	"errors"
	"io"
	"reflect"

	"github.com/beeker1121/goque"
)

// DequeueAll dequeues all elements of type dest from q
// and stores it into dest.
func DequeueAll(q Queue, dest interface{}) error {

	if dest == nil {
		return errors.New("destination cannot be nil")
	}

	typ := reflect.TypeOf(dest)

	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Slice {
		return errors.New("destination is not pointer to slice")
	}

	outKind := typ.Elem().Elem().Kind()

	for typ.Kind() == reflect.Ptr ||
		typ.Kind() == reflect.Interface ||
		typ.Kind() == reflect.Slice ||
		typ.Kind() == reflect.Array {

		typ = typ.Elem()
	}

	out := reflect.ValueOf(dest).Elem()

	for {
		vPtr := reflect.New(typ)

		err := q.Dequeue(vPtr.Interface())
		switch err {
		case nil:
		case io.EOF, goque.ErrEmpty, goque.ErrOutOfBounds:
			return nil
		default:
			return err
		}

		if outKind == reflect.Ptr {
			out.Set(reflect.Append(out, vPtr))
		} else {
			out.Set(reflect.Append(out, vPtr.Elem()))
		}
	}
}
