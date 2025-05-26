package yaml_encoder_test

import (
	"fmt"
	"strings"
	"testing"

	encoder "github.com/zwgblue/yaml-encoder"
)

func ExampleEncoder() {
	type DBConfig struct {
		Username string `comment:"this is the username of database"`
		Password string `comment:"this is the password of database"`
	}

	config := DBConfig{
		Username: "root",
		Password: "xxxxxx",
	}

	enc := encoder.NewEncoder(config, encoder.WithComments(encoder.CommentsOnHead))
	content, _ := enc.Encode()
	fmt.Printf("%s", content)
	// Output:
	// # this is the username of database
	// username: root
	// # this is the password of database
	// password: xxxxxx
}

func ExampleWithCustomizedTag() {
	type DBConfig struct {
		Username string `description:"this is the username of database"`
		Password string `description:"this is the password of database"`
	}

	config := DBConfig{
		Username: "root",
		Password: "xxxxxx",
	}

	enc := encoder.NewEncoder(config, encoder.WithComments(encoder.CommentsOnHead), encoder.WithCustomizedTag("description"))
	content, _ := enc.Encode()
	fmt.Printf("%s", content)
	// Output:
	// # this is the username of database
	// username: root
	// # this is the password of database
	// password: xxxxxx
}

type testCase struct {
	name     string
	input    interface{}
	options  []encoder.Option
	expected string
	wantErr  bool
}

func normalizeYAML(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func runTests(t *testing.T, tests []testCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := encoder.NewEncoder(tt.input, tt.options...)
			content, err := enc.Encode()

			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				actualNormalized := normalizeYAML(string(content))
				expectedNormalized := normalizeYAML(tt.expected)

				if actualNormalized != expectedNormalized {
					t.Errorf("Encode() output mismatch:\n----- Expected (%d bytes) -----\n%s\n----- Actual (%d bytes) -----\n%s\n----- End -----", len(expectedNormalized), expectedNormalized, len(actualNormalized), actualNormalized)
				}
			}
		})
	}
}

func TestOmitEmpty(t *testing.T) {
	type StructWithOmitEmpty struct {
		Name     string            `yaml:"name,omitempty"`
		Age      int               `yaml:"age,omitempty"`
		Count    *int              `yaml:"count,omitempty"`
		Tags     []string          `yaml:"tags,omitempty"`
		Attrs    map[string]string `yaml:"attrs,omitempty"`
		IsActive bool              `yaml:"isactive,omitempty"`
	}

	intZero := 0
	intVal := 10

	tests := []testCase{
		{
			name:     "all fields empty",
			input:    StructWithOmitEmpty{},
			expected: "{}",
		},
		{
			name:     "all fields empty with nil pointer, slice, map",
			input:    StructWithOmitEmpty{Count: nil, Tags: nil, Attrs: nil},
			expected: "{}",
		},
		{
			name:     "name not empty",
			input:    StructWithOmitEmpty{Name: "test"},
			expected: "name: test",
		},
		{
			name:     "age not empty",
			input:    StructWithOmitEmpty{Age: 10},
			expected: "age: 10",
		},
		{
			name:     "age is zero (should be omitted)",
			input:    StructWithOmitEmpty{Age: 0},
			expected: "{}",
		},
		{
			name:     "count not empty (pointer to non-zero)",
			input:    StructWithOmitEmpty{Count: &intVal},
			expected: "count: 10",
		},
		{
			name:     "count is pointer to zero",
			input:    StructWithOmitEmpty{Count: &intZero},
			expected: "count: 0",
		},
		{
			name:     "count is nil (should be omitted)",
			input:    StructWithOmitEmpty{Count: nil},
			expected: "{}",
		},
		{
			name:    "tags not empty", // Adjusted for 4-space indent from actual output
			input:   StructWithOmitEmpty{Tags: []string{"a"}},
			expected: "tags:\n    - a",
		},
		{
			name:     "tags empty (should be omitted)",
			input:    StructWithOmitEmpty{Tags: []string{}},
			expected: "{}",
		},
		{
			name:    "attrs not empty", // Adjusted for 4-space indent
			input:   StructWithOmitEmpty{Attrs: map[string]string{"key": "val"}},
			expected: "attrs:\n    key: val",
		},
		{
			name:     "attrs empty (should be omitted)",
			input:    StructWithOmitEmpty{Attrs: map[string]string{}},
			expected: "{}",
		},
		{
			name:     "isActive true",
			input:    StructWithOmitEmpty{IsActive: true},
			expected: "isactive: true",
		},
		{
			name:     "isActive false (zero value, should be omitted)",
			input:    StructWithOmitEmpty{IsActive: false},
			expected: "{}",
		},
		{
			name: "all fields non-empty", // Adjusted for 4-space indent
			input: StructWithOmitEmpty{
				Name:     "test",
				Age:      30,
				Count:    &intVal,
				Tags:     []string{"a", "b"},
				Attrs:    map[string]string{"k": "v"},
				IsActive: true,
			},
			expected: `
name: test
age: 30
count: 10
tags:
    - a
    - b
attrs:
    k: v
isactive: true`,
		},
		{
			name:    "omitempty disabled globally - all fields zero/nil",
			input:   StructWithOmitEmpty{},
			options: []encoder.Option{encoder.WithOmitEmpty(false)},
			expected: `
name: ""
age: 0
count: null
tags: []
attrs: {}
isactive: false`,
		},
		{
			name: "omitempty disabled globally - specific fields empty, Count points to zero",
			input:   StructWithOmitEmpty{Name: "hello", Age: 0, Count: &intZero, Tags: []string{}, Attrs: map[string]string{}},
			options: []encoder.Option{encoder.WithOmitEmpty(false)},
			expected: `
name: hello
age: 0
count: 0
tags: []
attrs: {}
isactive: false`,
		},
	}
	runTests(t, tests)
}

func TestComments(t *testing.T) {
	type SimpleStruct struct {
		Field1 string `yaml:"field1" comment:"Comment for Field1"`
		Field2 int    `yaml:"field2,omitempty" comment:"Comment for Field2 (omitempty)"`
	}
	type NestedStruct struct {
		Name      string        `yaml:"name" comment:"Comment for Name"`
		Simple    SimpleStruct  `yaml:"simple" comment:"Comment for SimpleStruct"`
		SimplePtr *SimpleStruct `yaml:"simplePtr,omitempty" comment:"Comment for SimplePtrStruct"`
	}

	tests := []testCase{
		{
			name:     "CommentsOnHead - SimpleStruct - Field1",
			input:    SimpleStruct{Field1: "val1"},
			options:  []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: "# Comment for Field1\nfield1: val1",
		},
		{
			name:    "CommentsOnHead - SimpleStruct - Field2 (not empty)",
			input:   SimpleStruct{Field1: "val1", Field2: 123},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Comment for Field1
field1: val1
# Comment for Field2 (omitempty)
field2: 123`,
		},
		{
			name:     "CommentsOnHead - SimpleStruct - Field2 (empty, omitted)",
			input:    SimpleStruct{Field1: "val1"},
			options:  []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: "# Comment for Field1\nfield1: val1",
		},
		{
			name: "CommentsOnHead - NestedStruct", // Adjusted for 4-space indent
			input: NestedStruct{
				Name:      "TestNested",
				Simple:    SimpleStruct{Field1: "s_val1", Field2: 10},
				SimplePtr: &SimpleStruct{Field1: "sp_val1", Field2: 20},
			},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Comment for Name
name: TestNested
# Comment for SimpleStruct
simple:
    # Comment for Field1
    field1: s_val1
    # Comment for Field2 (omitempty)
    field2: 10
# Comment for SimplePtrStruct
simplePtr:
    # Comment for Field1
    field1: sp_val1
    # Comment for Field2 (omitempty)
    field2: 20`,
		},
		{
			name:    "CommentsInLine - SimpleStruct",
			input:   SimpleStruct{Field1: "val1", Field2: 123},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsInLine)},
			expected: `
field1: val1 # Comment for Field1
field2: 123 # Comment for Field2 (omitempty)`,
		},
		{
			name:    "CommentsOnFoot - SimpleStruct", // Adjusted for actual yaml.v3 output
			input:   SimpleStruct{Field1: "val1", Field2: 123},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnFoot)},
			expected: `
field1: val1
# Comment for Field1

field2: 123
# Comment for Field2 (omitempty)`,
		},
		{
			name:     "CommentsOnHead - Empty Comment String",
			input:    struct{ Field string `yaml:"field" comment:""` }{Field: "value"},
			options:  []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: "field: value",
		},
		{
			name: "CommentsOnHead - Map Field", // Adjusted for 4-space indent
			input: struct { MyMap map[string]string `yaml:"myMap" comment:"Comment for MyMap"` }{
				MyMap: map[string]string{"key1": "val1"},
			},
			options:  []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: "# Comment for MyMap\nmyMap:\n    key1: val1",
		},
		{
			name: "CommentsOnHead - Slice Field", // Adjusted for 4-space indent
			input: struct { MySlice []string `yaml:"mySlice" comment:"Comment for MySlice"` }{
				MySlice: []string{"item1"},
			},
			options:  []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: "# Comment for MySlice\nmySlice:\n    - item1",
		},
	}
	runTests(t, tests)
}

func TestInlineFields(t *testing.T) {
	type InnerStruct struct {
		InnerField1 string `yaml:"innerField1" comment:"Inner 1"`
		InnerField2 int    `yaml:"innerField2,omitempty" comment:"Inner 2"`
	}
	type OuterStruct struct {
		OuterField string      `yaml:"outerField,omitempty" comment:"Outer"`
		Inner      InnerStruct `yaml:",inline"`
	}
	type OuterWithPtrInline struct {
		OuterField string      `yaml:"outerField,omitempty" comment:"Outer"`
		Inner      *InnerStruct `yaml:",inline"`
	}

	innerInstance := InnerStruct{InnerField1: "val1", InnerField2: 10}
	emptyInnerInstance := InnerStruct{}

	tests := []testCase{
		{
			name:    "Inline struct with values",
			input:   OuterStruct{OuterField: "outerVal", Inner: innerInstance},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Outer
outerField: outerVal
# Inner 1
innerField1: val1
# Inner 2
innerField2: 10`,
		},
		{
			name:    "Inline struct with omitempty on inner field (InnerField2 is zero)",
			input:   OuterStruct{OuterField: "outerVal", Inner: InnerStruct{InnerField1: "val1"}},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Outer
outerField: outerVal
# Inner 1
innerField1: val1`,
		},
		{
			name:    "Inline empty struct (InnerField1 is empty string, InnerField2 is zero and omitted)",
			input:   OuterStruct{OuterField: "outerVal", Inner: emptyInnerInstance},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Outer
outerField: outerVal
# Inner 1
innerField1: ""`,
		},
		{
			name:    "Inline nil pointer struct",
			input:   OuterWithPtrInline{OuterField: "outerVal", Inner: nil},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Outer
outerField: outerVal`,
		},
		{
			name:    "Inline non-nil pointer struct with values",
			input:   OuterWithPtrInline{OuterField: "outerVal", Inner: &innerInstance},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Outer
outerField: outerVal
# Inner 1
innerField1: val1
# Inner 2
innerField2: 10`,
		},
		{
			name:    "Inline non-nil pointer to empty struct",
			input:   OuterWithPtrInline{OuterField: "outerVal", Inner: &emptyInnerInstance},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Outer
outerField: outerVal
# Inner 1
innerField1: ""`,
		},
		{
			name:    "Only Inline struct",
			input:   struct{ Inner InnerStruct `yaml:",inline"` }{Inner: innerInstance},
			options: []encoder.Option{encoder.WithComments(encoder.CommentsOnHead)},
			expected: `
# Inner 1
innerField1: val1
# Inner 2
innerField2: 10`,
		},
	}
	runTests(t, tests)
}

func TestStructScenarios(t *testing.T) {
	type NoExportedFields struct {
		name string
		age  int
	}
	type MixedFields struct {
		Exported   string `yaml:"exported"`
		unexported string
		Another    int `yaml:"another"`
	}

	tests := []testCase{
		{
			name:     "Struct with no exported fields",
			input:    NoExportedFields{name: "test", age: 10},
			expected: "{}",
		},
		{
			name:     "Struct with no exported fields as pointer",
			input:    &NoExportedFields{name: "test", age: 10},
			expected: "{}",
		},
		{
			name:     "Struct with mixed exported and unexported fields",
			input:    MixedFields{Exported: "val1", unexported: "hidden", Another: 123},
			expected: "exported: val1\nanother: 123",
		},
		{
			name:     "Empty struct (no fields at all)",
			input:    struct{}{},
			expected: "{}",
		},
	}
	runTests(t, tests)
}

func TestMapScenarios(t *testing.T) {
	tests := []testCase{
		{
			name:  "Simple map string to string", // Adjusted for 4-space indent
			input: map[string]string{"key1": "val1", "b_key2": "val2", "a_key3": "val3"},
			expected: `
a_key3: val3
b_key2: val2
key1: val1`,
		},
		{
			name:  "Map string to int", // Adjusted for 4-space indent
			input: map[string]int{"one": 1, "two": 2},
			expected: `
one: 1
two: 2`,
		},
		{
			name:  "Map int to string (keys not quoted by yaml.v3)", // Adjusted for 4-space indent
			input: map[int]string{1: "one", 2: "two"},
			expected: `
1: one
2: two`,
		},
		{
			name:     "Empty map",
			input:    map[string]string{},
			expected: "{}",
		},
		{
			name:     "Nil map (marshals to empty map by yaml.v3)",
			input:    (map[string]string)(nil),
			expected: "{}",
		},
		{
			name: "Map with complex values (structs)", // Adjusted for 4-space indent
			input: map[string]struct { A string `yaml:"a"` }{
				"entry1": {A: "val_a"},
				"entry2": {A: "val_b"},
			},
			expected: `
entry1:
    a: val_a
entry2:
    a: val_b`,
		},
	}
	runTests(t, tests)
}

func TestSliceScenarios(t *testing.T) {
	type SimpleStructForSlice struct {
		X string `yaml:"x"`
	}
	tests := []testCase{
		{
			name:  "Slice of strings", // Adjusted for 4-space indent for list items from root
			input: []string{"a", "b", "c"},
			expected: `
- a
- b
- c`,
		},
		{
			name:  "Slice of ints", // Adjusted
			input: []int{1, 2, 3},
			expected: `
- 1
- 2
- 3`,
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: "[]",
		},
		{
			name:     "Nil slice (marshals to empty slice by yaml.v3)",
			input:    ([]string)(nil),
			expected: "[]",
		},
		{
			name:  "Slice of structs", // Adjusted
			input: []SimpleStructForSlice{{X: "one"}, {X: "two"}},
			expected: `
- x: one
- x: two`,
		},
		{
			name: "Slice of maps", // Adjusted
			input: []map[string]int{
				{"a": 1, "b": 2},
				{"c": 3},
			},
			expected: `
- a: 1
  b: 2
- c: 3`,
		},
		{
			name: "Slice with nil pointer element", // Adjusted
			input: []*SimpleStructForSlice{
				{X: "one"},
				nil,
				{X: "three"},
			},
			expected: `
- x: one
- null
- x: three`,
		},
	}
	runTests(t, tests)
}

type ErrorMarshaler struct{}
var customError = fmt.Errorf("custom MarshalYAML error")
func (em *ErrorMarshaler) MarshalYAML() (interface{}, error) { return nil, customError }

func TestErrorHandling(t *testing.T) {
	tests := []testCase{
		{
			name:    "Error from custom MarshalYAML",
			input:   &ErrorMarshaler{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := encoder.NewEncoder(tt.input, tt.options...)
			_, err := enc.Encode()

			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "custom MarshalYAML error") {
					 t.Errorf("Encode() error = %v, expected to contain %q", err, "custom MarshalYAML error")
				}
			}
		})
	}
}

type NilValMarshaler struct { Val string }
func (n *NilValMarshaler) MarshalYAML() (interface{}, error) {
	if n == nil { return nil, fmt.Errorf("MarshalYAML called on nil receiver") }
	return map[string]string{"val": n.Val}, nil
}

func TestSpecialCasesAndRefactoredHelpers(t *testing.T) {
	var nilMarshaler *NilValMarshaler = nil

	tests := []testCase{
		{
			name:     "yaml.Marshaler that is nil (pointer is nil)",
			input:    nilMarshaler,
			expected: "null",
			wantErr:  false,
		},
		{
			name:     "yaml.Marshaler that is not nil",
			input:    &NilValMarshaler{Val: "test"},
			expected: "val: test",
			wantErr:  false,
		},
	}
	runTests(t, tests)
}
