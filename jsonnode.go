package jsonnode

import "encoding/json"

var _ json.Marshaler = (*JSONNode)(nil)
var _ json.Unmarshaler = (*JSONNode)(nil)

// JSONNode represents a JSON node to be marshalled (TODO), or that has been unmarshalled.
// A JSONNode can be a whole JSON object that can be marshalled to JSON (TODO),
// or that has been unmarshalled from JSON.
// It can also represent a specific member (of any type) in a JSON object.
type JSONNode struct {
	parent    *JSONNode
	fieldName string
	data      map[string]interface{}
	index     int
}

// New creates a new JSONNode, ready to put data into to marshal to JSON
func New() *JSONNode {
	jn := new(JSONNode)
	jn.init()
	jn.data = make(map[string]interface{})

	return jn
}

func newChild(parent *JSONNode, fieldName string) *JSONNode {
	jn := new(JSONNode)
	jn.init()
	jn.parent = parent
	jn.fieldName = fieldName

	return jn
}

func (jn *JSONNode) init() {
	jn.parent = nil
	jn.fieldName = ""
	jn.data = nil
	jn.index = -1
}

// MarshalJSON marshals this instance to JSON
func (jn *JSONNode) MarshalJSON() ([]byte, error) {
	if jn == nil {
		return json.Marshal(nil)
	}

	return json.Marshal(jn.Value())
}

// UnmarshalJSON unmarshals JSON into this instance of JSONNode
func (jn *JSONNode) UnmarshalJSON(data []byte) error {
	jn.init()
	jn.data = make(map[string]interface{})

	return json.Unmarshal(data, &jn.data)
}

// Get gets specified child field of this JSONNode.
// If the field exists (even if it has no value), a *JSONNode will be returned.
// Otherwise, nil will be returned (including if this *JSONNode instance is nil).
func (jn *JSONNode) Get(fieldName string) *JSONNode {
	if jn == nil {
		return nil
	}

	switch t := jn.Value().(type) {
	case map[string]interface{}:
		// This node can have children
		if _, ok := t[fieldName]; !ok {
			// This node does not have this child
			return nil
		}

	default:
		// This node doesn't have child nodes
		return nil
	}

	child := newChild(jn, fieldName)

	return child
}

// Value gets the raw value of this node
func (jn *JSONNode) Value() interface{} {
	if jn == nil {
		return nil
	}

	if jn.data == nil && jn.parent != nil {
		// The actual value for this is in the parent (this is not the root node)
		val := jn.parent.Value()

		if jn.index >= 0 {
			// This node is an item in an array
			return val.([]interface{})[jn.index]
		}

		// This node is not an item in an array

		valMap := val.(map[string]interface{})

		if len(jn.fieldName) == 0 {
			// There's no field name specified (probably trying to get a struct in an array)
			return valMap
		}

		// This node is a field on a struct
		return valMap[jn.fieldName]
	}

	// The data is directly contained in this node (this is probably the root node)
	return jn.data
}

// ValueAsNode gets the value of a field as a *JSONNode.
// This is useful for when the value is a JSON struct in an array element.
func (jn *JSONNode) ValueAsNode() (*JSONNode, bool) {
	_, ok := jn.Value().(map[string]interface{})
	if !ok {
		return nil, false
	}

	node := newChild(jn, "")

	return node, true
}

// ValueAsString gets the value of the current node as string
func (jn *JSONNode) ValueAsString() (string, bool) {
	val, ok := jn.Value().(string)

	return val, ok
}

// ValueAsNumber gets the value of the current node as a number.
// Golangs stdlib will unmarshal any numeric JSON object as a float64, so that's what you get.
func (jn *JSONNode) ValueAsNumber() (float64, bool) {
	val, ok := jn.Value().(float64)

	return val, ok
}

// ValueAsSlice returns the value of the current node as a []*JSONNode.
func (jn *JSONNode) ValueAsSlice() ([]*JSONNode, bool) {
	val, ok := jn.Value().([]interface{})
	if !ok {
		return nil, false
	}

	nodes := make([]*JSONNode, len(val))

	for i := range val {
		node := newChild(jn, jn.fieldName)
		node.index = i

		nodes[i] = node
	}

	return nodes, true
}
