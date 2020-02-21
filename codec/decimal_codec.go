package codec

import (
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DecimalCodec is used to encode decimal.Decimal types in primitive.Decimal128 and
// decode primitive.Decimal128 in decimal.Decimal
type DecimalCodec struct {
}

// EncodeValue implements bsoncodec.ValueEncoder interface
func (dc *DecimalCodec) EncodeValue(ctx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if val.Type() != reflect.TypeOf(decimal.Decimal{}) {
		return bsoncodec.ValueEncoderError{
			Name:     "DecimalEncodeValue",
			Types:    []reflect.Type{reflect.TypeOf(decimal.Decimal{})},
			Received: val,
		}
	}
	ourDecimal := val.Interface().(decimal.Decimal)
	mongoDecimal, err := primitive.ParseDecimal128(ourDecimal.String())
	if err != nil {
		return fmt.Errorf("DecimalCodec: unable to convert decimal.Decimal to primitive.Decimal128 %v", err)
	}
	return vw.WriteDecimal128(mongoDecimal)
}

// DecodeValue implements bsoncodec.ValueEncoder interface
func (dc *DecimalCodec) DecodeValue(ctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != reflect.TypeOf(decimal.Decimal{}) {
		return bsoncodec.ValueDecoderError{
			Name:     "DecimalDecodeValue",
			Types:    []reflect.Type{reflect.TypeOf(decimal.Decimal{})},
			Received: val,
		}
	}
	if vr.Type() != bsontype.Decimal128 {
		return fmt.Errorf("cannot decode %v into a decimal.Decimal type", vr.Type())
	}
	mongoDecimal, err := vr.ReadDecimal128()
	if err != nil {
		return err
	}
	newDec, err := decimal.NewFromString(mongoDecimal.String())
	if err != nil {
		return fmt.Errorf("DecimalCodec: unable to convert mongo decimal to decimal.Decimal %v", err)
	}
	val.Set(reflect.ValueOf(newDec))
	return nil
}
