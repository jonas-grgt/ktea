package serdes

import (
	"bytes"
	"encoding/binary"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/assert"
	"ktea/sradmin"
	"testing"
)

func TestAvroDeserializer(t *testing.T) {
	schema := `
{
   "type" : "record",
   "namespace" : "ktea.test",
   "name" : "Person",
   "fields" : [
      { "name" : "Name" , "type" : "string" },
      { "name" : "Age" , "type" : "int" }
   ]
}
`

	t.Run("no data deserializes to empty string", func(t *testing.T) {
		deserializer := NewAvroDeserializer(sradmin.NewMock())

		res, err := deserializer.Deserialize(nil)

		assert.Nil(t, err)
		assert.Empty(t, res)

		res, err = deserializer.Deserialize([]byte{})

		assert.Nil(t, err)
		assert.Empty(t, res)
	})

	t.Run("deserialize", func(t *testing.T) {
		sraMock := sradmin.NewMock()
		sraMock.GetSchemaByIdFunc = func(id int) tea.Msg {

			return sradmin.SchemaByIdReceived{
				Schema: sradmin.Schema{
					Id:      "",
					Value:   schema,
					Version: 0,
					Err:     nil,
				},
			}
		}
		deserializer := NewAvroDeserializer(sraMock)

		codec, err := goavro.NewCodec(schema)
		if err != nil {
			return
		}
		data, err := codec.BinaryFromNative(nil, map[string]interface{}{
			"Name": "John",
			"Age":  21,
		})
		if err != nil {
			t.Error(err)
		}

		var buf bytes.Buffer
		buf.WriteByte(0x00)
		schemaID := int32(1)
		if err := binary.Write(&buf, binary.BigEndian, schemaID); err != nil {
			t.Error(err)
		}
		buf.Write(data)

		res, err := deserializer.Deserialize(buf.Bytes())
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, `{"Age":21,"Name":"John"}`, res.Value)
		assert.Equal(t, schema, res.Schema)
	})

	t.Run("deserialize failed", func(t *testing.T) {
		t.Run("invalid schema", func(t *testing.T) {
			sraMock := sradmin.NewMock()
			sraMock.GetSchemaByIdFunc = func(id int) tea.Msg {

				return sradmin.SchemaByIdReceived{
					Schema: sradmin.Schema{
						Id:      "",
						Value:   "{invalid}",
						Version: 0,
						Err:     nil,
					},
				}
			}
			deserializer := NewAvroDeserializer(sraMock)

			codec, err := goavro.NewCodec(schema)
			if err != nil {
				return
			}
			data, err := codec.BinaryFromNative(nil, map[string]interface{}{
				"Name": "John",
				"Age":  21,
			})
			if err != nil {
				t.Error(err)
			}

			var buf bytes.Buffer
			buf.WriteByte(0x00)
			schemaID := int32(1)
			if err := binary.Write(&buf, binary.BigEndian, schemaID); err != nil {
				t.Error(err)
			}
			buf.Write(data)

			res, err := deserializer.Deserialize(buf.Bytes())

			assert.Empty(t, res)
			assert.NotNil(t, err)
			assert.Equal(t, "cannot unmarshal schema JSON: invalid character 'i' looking for beginning of object key string", err.Error())
		})
	})
}
