package serdes

type Deserializer interface {
	Deserialize(data []byte) (DesData, error)
}
