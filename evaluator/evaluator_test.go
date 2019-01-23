package evaluator

import (
	"testing"

	"github.com/elsonwu/monkey-go/lexer"
	"github.com/elsonwu/monkey-go/object"
	"github.com/elsonwu/monkey-go/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer, got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}

	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},

		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != true", false},
		{"false != false", false},
		{"true != false", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean, got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}

	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if(true) {10}", 10},
		{"if(false) {10}", nil},
		{"if(1) {10}", 10},
		{"if(1 < 2) {10}", 10},
		{"if(1 > 2) {10}", nil},
		{"if(1 > 2) {10} else {20}", 20},
		{"if(1 < 2) {10} else {10}", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	return obj == NULL
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 8;", 10},
		{`
		if (10 > 1) {
			if (10 > 2) {
				return 10;
			}

			return 1;
		}
		
`, 10},
	}

	for _, tt := range tests {
		evalulated := testEval(tt.input)
		testIntegerObject(t, evalulated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input string

		expected string
	}{
		{"5 + true", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + true; 5;", "type mismatch: INTEGER + BOOLEAN"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "unknown operator: BOOLEAN + BOOLEAN"},
		{`
		if (10 > 1) {
			if (10 > 1) {
				return true + false;
			}

			return 1;
		}
`, "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "identifier not found: foobar"},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("[%d]no error object returned, got=%T(%+v)", i, evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expected {
			t.Errorf("wrong error message, expected=%q, got=%q", tt.expected, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := `fn(x) { x + 2;};`
	evaluated := testEval(input)

	fn, ok := evaluated.(*object.Function)

	if !ok {
		t.Fatalf("object is not Function, got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters, parameters=%+v", fn.Parameters)
	}

	expectedBody := `(x + 2)`

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q, got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let ident = fn(x) {x;}; ident(5);", 5},
		{"let ident = fn(x) { return x;}; ident(5);", 5},
		{"let ident = fn(x) {x*2;}; ident(5);", 10},
		{"let ident = fn(x, y) {x + y;}; ident(5, 5);", 10},
		{"let ident = fn(x, y) {x + y;}; ident(5 + 5, ident(5, 5));", 20},
		{"fn(x){x;}(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestBuiltFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments, got=2, want=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch ttType := tt.expected.(type) {
		case string:
			objErr, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if objErr.Message != tt.expected {
				t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, objErr.Message)
			}
		case int:
			testIntegerObject(t, evaluated, int64(ttType))
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	input := `[1, 2 * 3, 3 + 4]`
	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not object.Array, got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements, got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 6)
	testIntegerObject(t, result.Elements[2], 7)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][2]", 3},
		{"let i = 0; [1, 2, 3][i]", 1},
		{"[1, 2, 3][1 + 1]", 3},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		}
	}
}

func TestArrayPush(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let a = [1, 3, 5]; let b = push(a, 7); b;", "[1, 3, 5, 7]"},
		{"let a = [1, 3, 5]; let b = push(a, 7); a;", "[1, 3, 5]"},
		{`let a = [1, "hello", 5]; let b = push(a, "world", [9]); b;`, `[1, "hello", 5, "world", [9]]`},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("evaluated is not array, got=%T", evaluated)
		}

		if arr.Inspect() != tt.expected {
			t.Fatalf("evaluated want=%s, got=%s", tt.expected, arr.Inspect())
		}
	}
}

func TestHash(t *testing.T) {
	input := `
	let t = "two";
	{
		"one": 10 - 9,
		t: 1 + 1,
		"thr" + "ee": 6/2,
		4: 4,
		true: 5,
		false: 6
	}
	`
	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("evaluated is not *object.HashLiteral. got=%T", evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs, got=%d, want=%d", len(result.Pairs), len(expected))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in pairs")
		}

		pairValue, ok := pair.Value.(*object.Integer)
		if !ok {
			t.Errorf("pair value is not *object.Integer")
		}

		testIntegerObject(t, pairValue, expectedValue)
	}
}

func TestHashIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`{"a":1}["a"]`, 1},
		{`{"a":1,"b":2}["b"]`, 2},
		{`{"a":1+1,"b":2/2}["b"]`, 1},
		{`{1:1+1,"b":2/2}[1]`, 2},
		{`{1:1+1,1+1:2/2}[2]`, 1},
		{`{"a":1,true:2}[true]`, 2},
		{`{"a":1,false:2}[false]`, 2},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		result, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("evaluated is not *object.Integer. got=%T", evaluated)
		}

		testIntegerObject(t, result, tt.expected)
	}
}
