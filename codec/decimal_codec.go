package codec

import (
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type DecimalCodec struct{}

// EncodeValue implements bsoncodec.ValueEncoder interface
func (dc DecimalCodec) EncodeValue(ctx bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
	if val.Type() != reflect.TypeOf(decimal.Decimal{}) {
		return bson.ValueEncoderError{
			Name:     "DecimalEncodeValue",
			Types:    []reflect.Type{reflect.TypeOf(decimal.Decimal{})},
			Received: val,
		}
	}
	ourDecimal := val.Interface().(decimal.Decimal)
	mongoDecimal, err := decimalToPrimitive128(ourDecimal)
	if err != nil {
		return fmt.Errorf("DecimalCodec: unable to convert decimal.Decimal to primitive.Decimal128, %v", err)
	}
	return vw.WriteDecimal128(mongoDecimal)
}

// DecodeValue implements bsoncodec.ValueEncoder interface
func (dc DecimalCodec) DecodeValue(ctx bson.DecodeContext, vr bson.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != reflect.TypeOf(decimal.Decimal{}) {
		return bson.ValueDecoderError{
			Name:     "DecimalDecodeValue",
			Types:    []reflect.Type{reflect.TypeOf(decimal.Decimal{})},
			Received: val,
		}
	}

	var newDec decimal.Decimal

	switch vr.Type() {
	case bson.TypeDecimal128:
		mongoDecimal, err := vr.ReadDecimal128()
		if err != nil {
			return err
		}
		newDec, err = primitive128ToDecimal(mongoDecimal)
		if err != nil {
			return fmt.Errorf("DecimalCodec: unable to convert primitive.Decimal128 to decimal.Decimal %v", err)
		}
	case bson.TypeInt32:
		i32, err := vr.ReadInt32()
		if err != nil {
			return err
		}
		newDec = decimal.NewFromInt32(i32)
	case bson.TypeDouble:
		f64, err := vr.ReadDouble()
		if err != nil {
			return err
		}
		newDec = decimal.NewFromFloat(f64)
	default:
		return fmt.Errorf("cannot decode %v into a decimal.Decimal type", vr.Type())
	}

	val.Set(reflect.ValueOf(newDec))
	return nil
}

func decimalToPrimitive128(d decimal.Decimal) (bson.Decimal128, error) {
	coefficient, exp := d.Coefficient(), d.Exponent()
	mongoDecimal, ok := bson.ParseDecimal128FromBigInt(coefficient, int(exp))
	if !ok {
		return bson.Decimal128{}, fmt.Errorf("unable to parse Decimal128 from big int")
	}
	return mongoDecimal, nil
}

func primitive128ToDecimal(p bson.Decimal128) (decimal.Decimal, error) {
	bigInt, exp, err := p.BigInt()
	if err != nil {
		return decimal.Decimal{}, err
	}
	// convert exp to int32 should never be a problem since exp is in [-6176, 6111]
	d := decimal.NewFromBigInt(bigInt, int32(exp))
	return d, nil
}
