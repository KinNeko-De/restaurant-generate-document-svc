package protolua

import (
	"fmt"

	"github.com/kinneko-de/protobuf-go/encoding/protolua"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const KinNekoDeProtobufParentPackage = "kinnekode.protobuf"
const KinNekoDeProtobufDecimal = "Decimal"

type KinnekoDeProtobuf struct {
}

func (KinnekoDeProtobuf) Handle(fullName protoreflect.FullName) (protolua.MarshalFunc, error) {
	if fullName.Parent() == KinNekoDeProtobufParentPackage {
		switch fullName.Name() {
		case KinNekoDeProtobufDecimal:
			return convertDecimal, nil
		default:
			return nil, fmt.Errorf("%s is not supported yet", fullName.Name())
		}
	}
	return nil, nil
}

func convertDecimal(encodingRun protolua.EncodingRun, m protoreflect.Message) error {
	fd := m.Descriptor().Fields().ByNumber(1)
	val := m.Get(fd)
	encodingRun.Encoder.WriteNumber(val.String())
	return nil
}
