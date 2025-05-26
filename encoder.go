package yaml_encoder

import (
	"reflect"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Encoder implements config encoder.
type Encoder struct {
	value   interface{}
	options *Options
}

// CommentsFlag comments encoding flags type.
type CommentsFlag int

func (f CommentsFlag) enabled() bool {
	return f > CommentsDisabled
}

const (
	CommentsDisabled CommentsFlag = iota
	CommentsOnHead
	CommentsInLine
	CommentsOnFoot
)

// Options defines encoder config.
type Options struct {
	CommentsFlag   CommentsFlag
	OmitEmpty      bool
	CommentTagName string
}

func newOptions(opts ...Option) *Options {
	res := &Options{
		CommentsFlag:   CommentsDisabled,
		OmitEmpty:      true,
		CommentTagName: "comment", // default
	}

	for _, o := range opts {
		o(res)
	}

	return res
}

// Option gives ability to alter config encoder output settings.
type Option func(*Options)

// WithComments enables comments in the encoder.
func WithComments(flag CommentsFlag) Option {
	return func(o *Options) {
		o.CommentsFlag = flag
	}
}

// WithOmitEmpty toggles omitempty handling.
func WithOmitEmpty(value bool) Option {
	return func(o *Options) {
		o.OmitEmpty = value
	}
}

// WithCustomizedTag use customized tag for comment
func WithCustomizedTag(tagName string) Option {
	return func(o *Options) {
		if tagName != "" {
			o.CommentTagName = tagName
		}
	}
}

// NewEncoder initializes and returns an `Encoder`.
func NewEncoder(value interface{}, opts ...Option) *Encoder {
	return &Encoder{
		value:   value,
		options: newOptions(opts...),
	}
}

// Marshal converts value to YAML-serializable value (suitable for MarshalYAML).
func (e *Encoder) Marshal() (*yaml.Node, error) {
	node, err := toYamlNode(e.value, e.options)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// Encode converts value to yaml.
//nolint:gocyclo
func (e *Encoder) Encode() ([]byte, error) {
	if e.options.CommentsFlag == CommentsDisabled {
		return yaml.Marshal(e.value)
	}

	node, err := e.Marshal()
	if err != nil {
		return nil, err
	}

	// special handling for case when we get an empty output
	if handled, res, err := handleEmptyNodeComment(node, e.options); handled {
		return res, err
	}

	return yaml.Marshal(node)
}

// handleEmptyNodeComment handles the special case for empty mapping nodes with foot comments.
// It returns true if the case was handled, along with the marshalled bytes and any error.
func handleEmptyNodeComment(node *yaml.Node, options *Options) (bool, []byte, error) {
	if node.Kind == yaml.MappingNode && len(node.Content) == 0 && node.FootComment != "" && options.CommentsFlag.enabled() {
		res := ""

		if node.HeadComment != "" {
			res += node.HeadComment + "\n"
		}

		lines := strings.Split(res+node.FootComment, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
				continue
			}

			lines[i] = "# " + line
		}

		return true, []byte(strings.Join(lines, "\n")), nil
	}
	return false, nil, nil
}

func isEmpty(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Ptr:
		return value.IsNil()
	case reflect.Map:
		return len(value.MapKeys()) == 0
	case reflect.Slice:
		return value.Len() == 0
	default:
		return value.IsZero()
	}
}

func isNil(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

//nolint:gocyclo,cyclop
func toYamlNode(in interface{}, options *Options) (*yaml.Node, error) {
	node := &yaml.Node{}

	// flags := options.Comments

	// do not wrap yaml.Node into yaml.Node
	if n, ok := in.(*yaml.Node); ok {
		return n, nil
	}

	// if input implements yaml.Marshaler we should use that marshaller instead
	// same way as regular yaml marshal does
	if m, ok := in.(yaml.Marshaler); ok && !isNil(reflect.ValueOf(in)) {
		res, err := m.MarshalYAML()
		if err != nil {
			return nil, err
		}

		if n, ok := res.(*yaml.Node); ok {
			return n, nil
		}

		in = res
	}

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	//nolint:exhaustive
	switch v.Kind() {
	case reflect.Struct:
		node.Kind = yaml.MappingNode

		node.Kind = yaml.MappingNode
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if err := processStructField(node, v, t, i, options); err != nil {
				return nil, err
			}
		}
	case reflect.Map:
		node.Kind = yaml.MappingNode
		if err := processMap(v, node, options); err != nil {
			return nil, err
		}
	case reflect.Slice:
		node.Kind = yaml.SequenceNode
		if err := processSlice(v, node, options); err != nil {
			return nil, err
		}
	default:
		if err := node.Encode(in); err != nil {
			return nil, err
		}
	}

	return node, nil
}

// processStructField handles processing of a single struct field.
func processStructField(node *yaml.Node, v reflect.Value, t reflect.Type, i int, options *Options) error {
	// skip unexported fields
	if !v.Field(i).CanInterface() {
		return nil
	}

	field := t.Field(i)
	comment := field.Tag.Get(options.CommentTagName)
	tag := field.Tag.Get("yaml")
	parts := strings.Split(tag, ",")
	fieldName := parts[0]
	parts = parts[1:]

	if fieldName == "" {
		fieldName = strings.ToLower(field.Name)
	}

	if fieldName == "-" {
		return nil
	}

	var (
		fieldIsEmpty = isEmpty(v.Field(i)) // Determine if field is empty
		skipField    = false               // Default to not skipping the field
		inline       = false
		flow         = false
	)

	// Tag parsing logic
	hasOmitemptyTagInStruct := false
	for _, part := range parts {
		switch part {
		case "omitempty":
			hasOmitemptyTagInStruct = true
		case "inline":
			inline = true
		case "flow":
			flow = true
		}
	}

	// Decision logic for skipping the field
	performSkip := false // Initialize

	if fieldName == "-" {
		performSkip = true // Always skip if field name is "-"
	} else {
		// If options.OmitEmpty is true, then standard omitempty logic applies
		if options.OmitEmpty {
			if hasOmitemptyTagInStruct && fieldIsEmpty {
				performSkip = true
			}
		}
		// If options.OmitEmpty is false, performSkip remains false (unless fieldName was "-"),
		// meaning the omitempty tag is completely ignored because the condition
		// `hasOmitemptyTagInStruct && fieldIsEmpty` is only checked if options.OmitEmpty is true.
	}

	// This is the crucial assignment
	skipField = performSkip

	var value interface{}
	if v.Field(i).CanInterface() {
		value = v.Field(i).Interface()
	}

	if skipField {
		return nil
	}

	var style yaml.Style
	if flow {
		style |= yaml.FlowStyle
	}

	if inline {
		child, err := toYamlNode(value, options)
		if err != nil {
			return err
		}

		if child.Kind == yaml.MappingNode || child.Kind == yaml.SequenceNode {
			appendNodes(node, child.Content...)
		}
	} else if err := addToMap(node, fieldName, value, comment, style, options); err != nil {
		return err
	}
	return nil
}

func processMap(v reflect.Value, node *yaml.Node, options *Options) error {
	keys := v.MapKeys()
	// always interate keys in alphabetical order to preserve the same output for maps
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	for _, k := range keys {
		element := v.MapIndex(k)
		value := element.Interface()

		if err := addToMap(node, k.Interface(), value, "", 0, options); err != nil {
			return err
		}
	}
	return nil
}

func processSlice(v reflect.Value, node *yaml.Node, options *Options) error {
	nodes := make([]*yaml.Node, v.Len())
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i)
		var err error
		nodes[i], err = toYamlNode(element.Interface(), options)
		if err != nil {
			return err
		}
	}
	appendNodes(node, nodes...)
	return nil
}

func appendNodes(dest *yaml.Node, nodes ...*yaml.Node) {
	if dest.Content == nil {
		dest.Content = []*yaml.Node{}
	}

	dest.Content = append(dest.Content, nodes...)
}

func addToMap(dest *yaml.Node, fieldName, in interface{}, fieldComent string, style yaml.Style, options *Options) error {
	key, err := toYamlNode(fieldName, options)
	if err != nil {
		return err
	}

	value, err := toYamlNode(in, options)
	if err != nil {
		return err
	}
	value.Style = style

	if options.CommentsFlag.enabled() {
		addComment(key, fieldComent, options.CommentsFlag)
	}
	appendNodes(dest, key, value)

	return nil
}

func addComment(node *yaml.Node, comment string, flag CommentsFlag) {
	if !flag.enabled() || comment == "" { // Also explicitly skip if comment is empty
		return
	}

	switch flag {
	case CommentsOnHead:
		node.HeadComment = comment
	case CommentsInLine:
		node.LineComment = comment
	case CommentsOnFoot:
		node.FootComment = comment
	// No default needed as flag.enabled() already filters for valid comment placement flags.
	}
}
