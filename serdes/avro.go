package serdes

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
	"ktea/sradmin"
)

type GoAvroAvroDeserializer struct {
	subjects []sradmin.Subject
	sra      sradmin.SrAdmin
}

var ErrNoSchemaRegistry = errors.New("no schema registry configured")

func (d *GoAvroAvroDeserializer) Deserialize(data []byte) (string, error) {
	if data == nil || len(data) == 0 {
		return "", nil
	}

	schemaId, isAvro := isAvroWithSchemaID(data)

	if isAvro {

		if d.sra == nil {
			return "", fmt.Errorf("avro deserialization failed: %w", ErrNoSchemaRegistry)
		}

		if schema, err := d.getSchema(schemaId); err != nil {
			return "", err
		} else {
			var codec *goavro.Codec
			if codec, err = goavro.NewCodec(schema.Schema); err != nil {
				return "", err
			}

			deserData, _, err := codec.NativeFromBinary(data[5:])
			if err != nil {
				return "", err
			}

			jsonData, err := json.Marshal(deserData)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		}
	} else {
		return string(data), nil
	}

}

func (d *GoAvroAvroDeserializer) getSchema(schemaId int) (sradmin.Schema, error) {
	var schema sradmin.Schema

	switch msg := d.sra.GetSchemaById(schemaId).(type) {

	case sradmin.GettingSchemaByIdMsg:
		{
			switch msg := msg.AwaitCompletion().(type) {

			case sradmin.SchemaByIdReceived:
				{
					schema = msg.Schema
				}
			case sradmin.FailedToGetSchemaById:
				{
					return sradmin.Schema{}, msg.Err
				}
			}
		}

	case sradmin.SchemaByIdReceived:
		{
			schema = msg.Schema
		}
	}

	return schema, nil
}

func isAvroWithSchemaID(data []byte) (int, bool) {
	if len(data) < 5 {
		return -1, false
	}

	// Check the magic byte
	if data[0] != 0x00 {
		return -1, false
	}

	// Read the schema ID (4 bytes after the magic byte)
	var schemaId int32
	reader := bytes.NewReader(data[1:5])
	if err := binary.Read(reader, binary.BigEndian, &schemaId); err != nil {
		return -1, false
	}

	return int(schemaId), true
}

func NewAvroDeserializer(sra sradmin.SrAdmin) Deserializer {
	return &GoAvroAvroDeserializer{sra: sra}
}
