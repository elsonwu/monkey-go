package evaluator

import (
	"fmt"

	"github.com/elsonwu/monkey-go/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) > 1 {
				return newError("wrong number of arguments, got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) > 1 {
				return newError("wrong number of arguments, got=%d, want=1", len(args))
			}

			arg, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `first` not supported, got %s", args[0].Type())
			}

			if len(arg.Elements) > 0 {
				return arg.Elements[0]
			}

			return NULL
		},
	},
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) > 1 {
				return newError("wrong number of arguments, got=%d, want=1", len(args))
			}

			arg, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `rest` not supported, got %s", args[0].Type())
			}

			l := len(arg.Elements)
			if l == 0 {
				return NULL
			}

			newArr := make([]object.Object, l-1, l-1)
			copy(newArr, arg.Elements[1:l])
			return &object.Array{Elements: newArr}
		},
	},
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) > 1 {
				return newError("wrong number of arguments, got=%d, want=1", len(args))
			}

			arg, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `last` not supported, got %s", args[0].Type())
			}

			l := len(arg.Elements)
			if l == 0 {
				return NULL
			}

			return arg.Elements[l-1]
		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments, got=%d, want=2+", len(args))
			}

			arg, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `push` not supported, want ARRAY, got %s", args[0].Type())
			}

			l := len(arg.Elements)
			newElements := make([]object.Object, l, l+len(args[1:]))
			copy(newElements, arg.Elements)
			newElements = append(newElements, args[1:]...)
			return &object.Array{Elements: newElements}
		},
	},
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments, got=%d, want=1+", len(args))
			}

			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
}
