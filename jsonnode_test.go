package jsonnode

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const nestedJSON string = `{
    "status": "ok",
    "installation": {
        "license_key": null,
        "entitlement": null,
        "product": {
            "id": 103,
            "name": "Nebula Cloud Console Account",
            "customer_type": "business",
            "active": true,
            "code": "NCCA-B",
            "two_char_code": "NC",
            "grace_multiplier": 1.0,
            "grace_term_days": 0,
            "is_trial_allowed": false,
            "trial_duration": 30,
            "trial_max_volume": 0,
            "default_key_type": "keystone",
            "parent_id": null,
            "allow_grace": false,
            "enforce_volume": true,
            "sellable": true,
            "real_product_codes": null,
            "created_at": "2017-02-07T18:08:32Z",
            "updated_at": "2017-02-07T18:08:32Z"
        },
        "installation_token": "3SGWHsMh6Sxtcovhgvsz1487266817",
        "trial_status": "trial_available",
        "trial_max_volume": 0,
        "trial_starts_on": null,
        "trial_ends_on": null,
        "machine_id": "b4aa12b3-4110-4966-8f4c-ef8ff298d613",
        "product_id": 103,
        "notes": null,
        "product_version": "1.0.0",
        "product_build": null,
        "ip_address": "207.98.208.136, 207.98.208.136",
        "volume_used": 1,
        "last_contacted_at": "2017-02-16T17:40:17Z",
        "registered_at": "2017-02-16T17:40:17Z",
        "redeemed_at": null
    }
}`

func TestJSONNode(t *testing.T) {
	t.Parallel()

	toMarshal := make(map[string]interface{})
	toMarshal["int32"] = int32(-12345)
	toMarshal["uint32"] = uint32(12345)
	toMarshal["float32"] = float32(-123.45)

	when, err := time.Parse(time.RFC3339, "2017-06-28T18:00:00-04:00")
	require.NoError(t, err)

	toMarshal["time"] = when

	jsonBytes, err := json.Marshal(toMarshal)
	require.NoError(t, err)

	t.Logf("Marshalled: %s", string(jsonBytes))

	t.Run("stdlib example", func(t *testing.T) {
		// This test is just to demonstrate how different JSON types get unmarshalled into golang types.
		// Numeric types all get unmarshalled as float64, for example. Time just as a string.

		t.Parallel()

		var unmarshalled map[string]interface{}
		err = json.Unmarshal(jsonBytes, &unmarshalled)
		require.NoError(t, err)

		t.Logf("Unmarshalled (%T): %+[1]v", unmarshalled)

		for k, v := range unmarshalled {
			t.Logf(`%q (%T): %+[2]v`, k, v)
		}
	})

	t.Run("handle nil politely", func(t *testing.T) {
		t.Parallel()

		var jn *JSONNode

		// Make sure trying to call Get on a nil instance doesn't panic
		require.NotPanics(t, func() {
			jn.Get("does not exist")
		})

		// Make sure trying to call Value on a nil instance doesn't panic
		require.NotPanics(t, func() {
			val := jn.Value()
			require.Nil(t, val)
		})

		// Now with an actual instance

		jn = new(JSONNode)
		err = json.Unmarshal(jsonBytes, jn)
		require.NoError(t, err)

		// Non-existant field on the root of the object
		node := jn.Get("does not exist")
		require.Nil(t, node)
	})

	t.Run("simple unmarshal", func(t *testing.T) {
		t.Parallel()

		jn := new(JSONNode)
		err = json.Unmarshal(jsonBytes, jn)
		require.NoError(t, err)
		require.NotNil(t, jn.data)

		node := jn.Get("int32")
		require.NotNil(t, node)

		valFloat64, ok := node.ValueAsNumber()
		require.True(t, ok)
		require.Equal(t, float64(-12345), valFloat64)

		// It isn't a string.
		valString, ok := node.ValueAsString()
		require.False(t, ok)
		require.Zero(t, valString)

		// It isn't a slice.
		valSlice, ok := node.ValueAsSlice()
		require.False(t, ok)
		require.Zero(t, valSlice)

		// It isn't a node that can have children.
		valNode, ok := node.ValueAsNode()
		require.False(t, ok)
		require.Nil(t, valNode)

		// Non-existant field on a field that cannot have children
		node = node.Get("does not exist")
		require.Nil(t, node)

		valFloat64, ok = jn.Get("uint32").ValueAsNumber()
		require.True(t, ok)
		require.Equal(t, float64(12345), valFloat64)

		valFloat64, ok = jn.Get("float32").ValueAsNumber()
		require.True(t, ok)
		require.Equal(t, float64(-123.45), valFloat64)

		// JSON just handles time values as strings
		valString, ok = jn.Get("time").ValueAsString()
		require.True(t, ok)
		require.Equal(t, "2017-06-28T18:00:00-04:00", valString)
	})

	raw := `{
    "platter": "slate",
    "cheeses": ["cheddar", "swiss", "manchego"],
    "with": {
        "fruit": [{
                "type": "grapes",
                "count": 8
            },
            {
                "type": "strawberries",
                "count": 3
            }
        ],
        "meat": "prosciutto"
    }
}`

	t.Run("nested unmarshal", func(t *testing.T) {
		t.Parallel()

		jn := new(JSONNode)
		err = json.Unmarshal([]byte(raw), jn)
		require.NoError(t, err)
		t.Logf("Unmarshalled: %#v", jn)
		require.NotNil(t, jn.data)

		node0 := jn.Get("with")
		require.NotNil(t, node0)

		// A child node
		node1 := node0.Get("meat")
		require.NotNil(t, node1)

		// Value of the child node
		val := node1.Value()
		require.NotNil(t, val)
		require.Equal(t, "prosciutto", val)
	})

	t.Run("array of simple type", func(t *testing.T) {
		t.Parallel()

		jn := new(JSONNode)
		err := json.Unmarshal([]byte(raw), jn)
		require.NoError(t, err)

		cheeses := jn.Get("cheeses")
		require.NotNil(t, cheeses)

		nodes, ok := cheeses.ValueAsSlice()
		require.True(t, ok)
		require.Len(t, nodes, 3)

		for i, cheese := range nodes {
			chs, ok := cheese.ValueAsString()
			require.True(t, ok)

			t.Logf("%d: %q", i, chs)

			switch i {
			case 0:
				require.Equal(t, "cheddar", chs)

			case 1:
				require.Equal(t, "swiss", chs)

			case 2:
				require.Equal(t, "manchego", chs)
			}
		}
	})

	t.Run("array of struct", func(t *testing.T) {
		t.Parallel()

		jn := new(JSONNode)
		err := json.Unmarshal([]byte(raw), jn)
		require.NoError(t, err)

		fruit := jn.Get("with").Get("fruit")
		require.NotNil(t, fruit)

		fruitSlice, ok := fruit.ValueAsSlice()
		require.True(t, ok)
		require.Len(t, fruitSlice, 2)

		for i := range fruitSlice {
			fruitNode, ok := fruitSlice[i].ValueAsNode()
			require.True(t, ok)

			fruitType := fruitNode.Get("type")
			require.NotNil(t, fruitType)

			fruitTypeVal, ok := fruitType.ValueAsString()
			require.True(t, ok)

			count := fruitNode.Get("count")
			require.NotNil(t, count)

			countVal, ok := count.ValueAsNumber()
			require.True(t, ok)

			switch i {
			case 0:
				require.Equal(t, "grapes", fruitTypeVal)
				require.Equal(t, float64(8), countVal)

			case 1:
				require.Equal(t, "strawberries", fruitTypeVal)
				require.Equal(t, float64(3), countVal)
			}
		}
	})
}

func ExampleJSONNode_thorough() {
	raw := `{
    "platter": "slate",
    "cheeses": ["cheddar", "swiss", "manchego"],
    "with": {
        "fruit": [{
                "type": "grapes",
                "count": 8
            },
            {
                "type": "strawberries",
                "count": 3
            }
        ],
        "meat": "prosciutto"
    }
}`

	jn := new(JSONNode)
	err := json.Unmarshal([]byte(raw), jn)
	if err != nil {
		panic(err)
	}

	platter := jn.Get("platter")
	if platter == nil {
		panic(`The "patter" element doesn't exist`)
	}

	platterVal, ok := platter.ValueAsString()
	if !ok {
		panic(`The "patter" element isn't a string`)
	}

	fmt.Printf("Platter: %s\n", platterVal)

	cheeses := jn.Get("cheeses")
	if cheeses == nil {
		// No cheeses. 😢
		panic(`The "cheeses" element doesn't exist`)
	}

	chessesVal, ok := cheeses.ValueAsSlice()
	if !ok {
		// No slice of cheese
		// (☞ﾟヮﾟ)☞
		// ☜(ﾟヮﾟ☜)
		panic(`The "cheeses" element isn't a JSON array`)
	}

	fmt.Printf("Cheeses (%d):\n", len(chessesVal))

	for i := range chessesVal {
		cheese, ok := chessesVal[i].ValueAsString()
		if !ok {
			// This isn't cheese.
			panic(fmt.Sprintf(`Item %d in "cheeses" JSON array is not a string`, i))
		}

		fmt.Printf("    %s\n", cheese)
	}

	with := jn.Get("with")
	if with == nil {
		// Nothing with the cheese.
		panic(`Element "with" does not exist`)
	}

	meat := with.Get("meat")
	if meat == nil {
		panic(`Element "meat" does not exist`)
	}

	meatVal, ok := meat.ValueAsString()
	if !ok {
		panic(`Element "with.meat" is not a string`)
	}

	fmt.Printf("Meat: %s\n", meatVal)

	fruit := with.Get("fruit")
	if fruit == nil {
		panic(`Element "with.fruit" does not exist`)
	}

	fruitVal, ok := fruit.ValueAsSlice()
	if !ok {
		panic(`Element "with.fruit" is not a JSON array`)
	}

	fmt.Printf("Fruit (%d):\n", len(fruitVal))

	for i := range fruitVal {
		fruitNode, ok := fruitVal[i].ValueAsNode()
		if !ok {
			panic(fmt.Sprintf(`"with.fruit[%d]" is not a JSON object`, i))
		}

		fruitType := fruitNode.Get("type")
		if fruitType == nil {
			panic(fmt.Sprintf(`"with.fruit[%d].type" does not exist`, i))
		}

		f, ok := fruitType.ValueAsString()
		if !ok {
			panic(fmt.Sprintf(`"with.fruit[%d].type" is not a string`, i))
		}

		fmt.Printf("    %s\n", f)
	}

	// Output:
	// Platter: slate
	// Cheeses (3):
	//     cheddar
	//     swiss
	//     manchego
	// Meat: prosciutto
	// Fruit (2):
	//     grapes
	//     strawberries
}

func ExampleJSONNode_simple() {
	raw := `{
    "platter": "slate",
    "cheeses": ["cheddar", "swiss", "manchego"],
    "with": {
        "fruit": [{
                "type": "grapes",
                "count": 8
            },
            {
                "type": "strawberries",
                "count": 3
            }
        ],
        "meat": "prosciutto"
    }
}`

	jn := new(JSONNode)
	err := json.Unmarshal([]byte(raw), jn)
	if err != nil {
		panic(err)
	}

	platter, ok := jn.Get("platter").ValueAsString()
	if !ok {
		panic("The patter value isn't a string")
	}

	fmt.Printf("Platter: %s\n", platter)

	cheeses, ok := jn.Get("cheeses").ValueAsSlice()
	if !ok {
		// No slice of cheese
		// (☞ﾟヮﾟ)☞
		// ☜(ﾟヮﾟ☜)
		panic(`The "cheeses" element does not exist or is not a JSON array`)
	}

	fmt.Printf("Cheeses (%d):\n", len(cheeses))

	for i := range cheeses {
		cheese, ok := cheeses[i].ValueAsString()
		if !ok {
			// This isn't cheese.
			panic(fmt.Sprintf(`Item %d in "cheeses" JSON array is not a string`, i))
		}

		fmt.Printf("    %s\n", cheese)
	}

	meat, ok := jn.Get("with").Get("meat").ValueAsString()
	if !ok {
		panic(`Element "with.meat" does not exist or is not a string`)
	}

	fmt.Printf("Meat: %s\n", meat)

	fruit, ok := jn.Get("with").Get("fruit").ValueAsSlice()
	if !ok {
		panic(`Element "with.fruit" does not exist or is not a JSON array`)
	}

	fmt.Printf("Fruit (%d):\n", len(fruit))

	for i := range fruit {
		fruitNode, ok := fruit[i].ValueAsNode()
		if !ok {
			panic(fmt.Sprintf(`"with.fruit[%d]" is not a JSON object`, i))
		}

		fruitType, ok := fruitNode.Get("type").ValueAsString()
		if !ok {
			panic(fmt.Sprintf(`"with.fruit[%d].type" does not exist or is not a string`, i))
		}

		fmt.Printf("    %s\n", fruitType)
	}

	// Output:
	// Platter: slate
	// Cheeses (3):
	//     cheddar
	//     swiss
	//     manchego
	// Meat: prosciutto
	// Fruit (2):
	//     grapes
	//     strawberries
}

func ExampleJSONNode_ValueAsNode() {
	raw := `{
    "platter": "slate",
    "cheeses": ["cheddar", "swiss", "manchego"],
    "with": {
        "fruit": [{
                "type": "grapes",
                "count": 8
            },
            {
                "type": "strawberries",
                "count": 3
            }
        ],
        "meat": "prosciutto"
    }
}`

	jn := new(JSONNode)
	err := json.Unmarshal([]byte(raw), jn)
	if err != nil {
		panic(err)
	}

	// Get the "with" member as a *JSONNode
	with, ok := jn.Get("with").ValueAsNode()
	if !ok {
		panic(`No "with" member or its value is not a JavaScript object`)
	}

	// Get the "with.fruit" member as a slice of *JSONNode
	fruit, ok := with.Get("fruit").ValueAsSlice()
	if !ok {
		panic(`No "fruit" member or its value is not a JavaScript array`)
	}

	fmt.Printf("Fruit (%d):\n", len(fruit))

	for _, child := range fruit {
		// Get the "count" member of this element in the "fruit" JSON array
		count, ok := child.Get("count").ValueAsNumber()
		if !ok {
			panic(`No "count" member or its value is not a JavaScript number`)
		}

		// Get the "type" of this element in the "fruit" JSON array
		fruitType, ok := child.Get("type").ValueAsString()
		if !ok {
			panic(`No "type" member or its value is not a JavaScript string`)
		}

		fmt.Printf("    %.0f %s\n", count, fruitType)
	}

	// Output:
	// Fruit (2):
	//     8 grapes
	//     3 strawberries
}
