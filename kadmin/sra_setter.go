package kadmin

import (
	"ktea/sradmin"
)

type SraSetter interface {
	SetSra(sra sradmin.Client)
}

func (ka *SaramaKafkaAdmin) SetSra(sra sradmin.Client) {
	ka.sra = sra
}
